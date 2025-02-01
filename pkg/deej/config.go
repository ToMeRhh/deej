package deej

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/tomerhh/deej/pkg/deej/util"
)

// CanonicalConfig provides application-wide access to configuration fields,
// as well as loading/file watching logic for deej's configuration file
type CanonicalConfig struct {
	SliderMapping                *sliderMap
	MuteButtonMapping            *sliderMap
	AvailableOutputDeviceMapping *sliderMap

	UdpConnectionInfo struct {
		UdpPort int
	}
	TcpConnectionInfo struct {
		TcpPort int
	}

	InvertSliders bool

	NoiseReductionLevel string

	logger             *zap.SugaredLogger
	notifier           Notifier
	stopWatcherChannel chan bool

	reloadConsumers []chan bool

	userConfig     *viper.Viper
	internalConfig *viper.Viper
}

const (
	userConfigFilepath     = "config.yaml"
	internalConfigFilepath = "preferences.yaml"

	userConfigName     = "config"
	internalConfigName = "preferences"

	userConfigPath = "."

	configType = "yaml"

	configKeySliderMapping                = "slider_mapping"
	configKeyMuteButtonMapping            = "mute_button_mapping"
	configKeyAvailableOutputDeviceMapping = "available_output_device"
	configKeyInvertSliders                = "invert_sliders"
	configKeyNoiseReductionLevel          = "noise_reduction"
	configKeyUdpPort                      = "udp_port"
	configKeyTcpPort                      = "tcp_port"

	defaultUdpPort = 16990
	defaultTcpPort = 16991
)

// has to be defined as a non-constant because we're using path.Join
var internalConfigPath = path.Join(".", logDirectory)

// NewConfig creates a config instance for the deej object and sets up viper instances for deej's config files
func NewConfig(logger *zap.SugaredLogger, notifier Notifier) (*CanonicalConfig, error) {
	logger = logger.Named("config")

	cc := &CanonicalConfig{
		logger:             logger,
		notifier:           notifier,
		reloadConsumers:    []chan bool{},
		stopWatcherChannel: make(chan bool),
	}

	// distinguish between the user-provided config (config.yaml) and the internal config (logs/preferences.yaml)
	userConfig := viper.New()
	userConfig.SetConfigName(userConfigName)
	userConfig.SetConfigType(configType)
	userConfig.AddConfigPath(userConfigPath)

	userConfig.SetDefault(configKeySliderMapping, map[string][]string{})
	userConfig.SetDefault(configKeyMuteButtonMapping, map[string][]string{})
	userConfig.SetDefault(configKeyAvailableOutputDeviceMapping, map[string][]string{})
	userConfig.SetDefault(configKeyInvertSliders, false)

	userConfig.SetDefault(configKeyUdpPort, defaultUdpPort)
	userConfig.SetDefault(configKeyTcpPort, defaultTcpPort)

	internalConfig := viper.New()
	internalConfig.SetConfigName(internalConfigName)
	internalConfig.SetConfigType(configType)
	internalConfig.AddConfigPath(internalConfigPath)

	cc.userConfig = userConfig
	cc.internalConfig = internalConfig

	logger.Debug("Created config instance")

	return cc, nil
}

// Load reads deej's config files from disk and tries to parse them
func (cc *CanonicalConfig) Load() error {
	cc.logger.Debugw("Loading config", "path", userConfigFilepath)

	// make sure it exists
	if !util.FileExists(userConfigFilepath) {
		cc.logger.Warnw("Config file not found", "path", userConfigFilepath)
		cc.notifier.Notify("Can't find configuration!",
			fmt.Sprintf("%s must be in the same directory as deej. Please re-launch", userConfigFilepath))

		return fmt.Errorf("config file doesn't exist: %s", userConfigFilepath)
	}

	// load the user config
	if err := cc.userConfig.ReadInConfig(); err != nil {
		cc.logger.Warnw("Viper failed to read user config", "error", err)

		// if the error is yaml-format-related, show a sensible error. otherwise, show 'em to the logs
		if strings.Contains(err.Error(), "yaml:") {
			cc.notifier.Notify("Invalid configuration!",
				fmt.Sprintf("Please make sure %s is in a valid YAML format.", userConfigFilepath))
		} else {
			cc.notifier.Notify("Error loading configuration!", "Please check deej's logs for more details.")
		}

		return fmt.Errorf("read user config: %w", err)
	}

	// load the internal config - this doesn't have to exist, so it can error
	if err := cc.internalConfig.ReadInConfig(); err != nil {
		cc.logger.Debugw("Viper failed to read internal config", "error", err, "reminder", "this is fine")
	}

	// canonize the configuration with viper's helpers
	if err := cc.populateFromVipers(); err != nil {
		cc.logger.Warnw("Failed to populate config fields", "error", err)
		return fmt.Errorf("populate config fields: %w", err)
	}

	cc.logger.Info("Loaded config successfully")
	cc.logger.Infow("Config values",
		"sliderMapping", cc.SliderMapping,
		"muteButtonMapping", cc.MuteButtonMapping,
		"availableOutputDeviceMapping", cc.AvailableOutputDeviceMapping,
		"udpConnectionInfo", cc.UdpConnectionInfo,
		"invertSliders", cc.InvertSliders)

	return nil
}

// SubscribeToChanges allows external components to receive updates when the config is reloaded
func (cc *CanonicalConfig) SubscribeToChanges() chan bool {
	c := make(chan bool)
	cc.reloadConsumers = append(cc.reloadConsumers, c)

	return c
}

// WatchConfigFileChanges starts watching for configuration file changes
// and attempts reloading the config when they happen
func (cc *CanonicalConfig) WatchConfigFileChanges() {
	cc.logger.Debugw("Starting to watch user config file for changes", "path", userConfigFilepath)

	const (
		minTimeBetweenReloadAttempts = time.Millisecond * 500
		delayBetweenEventAndReload   = time.Millisecond * 50
	)

	lastAttemptedReload := time.Now()

	// establish watch using viper as opposed to doing it ourselves, though our internal cooldown is still required
	cc.userConfig.WatchConfig()
	cc.userConfig.OnConfigChange(func(event fsnotify.Event) {

		// when we get a write event...
		if event.Op&fsnotify.Write == fsnotify.Write {

			now := time.Now()

			// ... check if it's not a duplicate (many editors will write to a file twice)
			if lastAttemptedReload.Add(minTimeBetweenReloadAttempts).Before(now) {

				// and attempt reload if appropriate
				cc.logger.Debugw("Config file modified, attempting reload", "event", event)

				// wait a bit to let the editor actually flush the new file contents to disk
				<-time.After(delayBetweenEventAndReload)

				if err := cc.Load(); err != nil {
					cc.logger.Warnw("Failed to reload config file", "error", err)
				} else {
					cc.logger.Info("Reloaded config successfully")
					cc.notifier.Notify("Configuration reloaded!", "Your changes have been applied.")

					cc.onConfigReloaded()
				}

				// don't forget to update the time
				lastAttemptedReload = now
			}
		}
	})

	// wait till they stop us
	<-cc.stopWatcherChannel
	cc.logger.Debug("Stopping user config file watcher")
	cc.userConfig.OnConfigChange(nil)
}

// StopWatchingConfigFile signals our filesystem watcher to stop
func (cc *CanonicalConfig) StopWatchingConfigFile() {
	cc.stopWatcherChannel <- true
}

func (cc *CanonicalConfig) populateFromVipers() error {

	// merge the slider mappings from the user and internal configs
	cc.SliderMapping = sliderMapFromConfigs(
		cc.userConfig.GetStringMapStringSlice(configKeySliderMapping),
		cc.internalConfig.GetStringMapStringSlice(configKeySliderMapping),
	)
	// merge the slider mappings from the user and internal configs
	cc.MuteButtonMapping = sliderMapFromConfigs(
		cc.userConfig.GetStringMapStringSlice(configKeyMuteButtonMapping),
		cc.internalConfig.GetStringMapStringSlice(configKeyMuteButtonMapping),
	)
	// merge the slider mappings from the user and internal configs
	cc.AvailableOutputDeviceMapping = sliderMapFromConfigs(
		cc.userConfig.GetStringMapStringSlice(configKeyAvailableOutputDeviceMapping),
		cc.internalConfig.GetStringMapStringSlice(configKeyAvailableOutputDeviceMapping),
	)

	// get the rest of the config fields - viper saves us a lot of effort here

	cc.UdpConnectionInfo.UdpPort = cc.userConfig.GetInt(configKeyUdpPort)
	if (cc.UdpConnectionInfo.UdpPort <= 0) || (cc.UdpConnectionInfo.UdpPort >= 65536) {
		cc.logger.Warnw("Invalid UDP port specified, using default value",
			"key", configKeyUdpPort,
			"invalidValue", cc.UdpConnectionInfo.UdpPort,
			"defaultValue", defaultUdpPort)

		cc.UdpConnectionInfo.UdpPort = defaultUdpPort
	}
	cc.TcpConnectionInfo.TcpPort = cc.userConfig.GetInt(configKeyTcpPort)
	if (cc.TcpConnectionInfo.TcpPort <= 0) || (cc.TcpConnectionInfo.TcpPort >= 65536) {
		cc.logger.Warnw("Invalid TCP port specified, using default value",
			"key", configKeyTcpPort,
			"invalidValue", cc.TcpConnectionInfo.TcpPort,
			"defaultValue", defaultTcpPort)

		cc.TcpConnectionInfo.TcpPort = defaultTcpPort
	}

	cc.InvertSliders = cc.userConfig.GetBool(configKeyInvertSliders)
	cc.NoiseReductionLevel = cc.userConfig.GetString(configKeyNoiseReductionLevel)

	cc.logger.Debug("Populated config fields from vipers")

	return nil
}

func (cc *CanonicalConfig) onConfigReloaded() {
	cc.logger.Debug("Notifying consumers about configuration reload")

	for _, consumer := range cc.reloadConsumers {
		consumer <- true
	}
}
