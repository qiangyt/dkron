package dkron

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"time"

	flag "github.com/spf13/pflag"
)

// Config stores all configuration options for the dkron package.
type Config struct {
	NodeName              string `mapstructure:"node-name"`
	BindAddr              string `mapstructure:"bind-addr"`
	HTTPAddr              string `mapstructure:"http-addr"`
	Discover              string
	Backend               string
	BackendMachines       []string `mapstructure:"backend-machine"`
	Profile               string
	Interface             string
	AdvertiseAddr         string            `mapstructure:"advertise-addr"`
	Tags                  map[string]string `mapstructure:"tags"`
	SnapshotPath          string            `mapstructure:"snapshot-path"`
	ReconnectInterval     time.Duration     `mapstructure:"reconnect-interval"`
	ReconnectTimeout      time.Duration     `mapstructure:"reconnect-timeout"`
	TombstoneTimeout      time.Duration     `mapstructure:"tombstone-timeout"`
	DisableNameResolution bool              `mapstructure:"disable-name-resolution"`
	KeyringFile           string            `mapstructure:"keyring-file"`
	RejoinAfterLeave      bool              `mapstructure:"rejoin-after-leave"`
	Server                bool
	EncryptKey            string   `mapstructure:"encrypt-key"`
	StartJoin             []string `mapstructure:"start-join"`
	Keyspace              string
	RPCPort               int    `mapstructure:"rpc-port"`
	AdvertiseRPCPort      int    `mapstructure:"advertise-rpc-port"`
	LogLevel              string `mapstructure:"log-level"`

	MailHost          string `mapstructure:"mail-host"`
	MailPort          uint16 `mapstructure:"mail-port"`
	MailUsername      string `mapstructure:"mail-username"`
	MailPassword      string `mapstructure:"mail-password"`
	MailFrom          string `mapstructure:"mail-from"`
	MailPayload       string `mapstructure:"mail-payload"`
	MailSubjectPrefix string `mapstructure:"mail-subject-prefix"`

	WebhookURL     string   `mapstructure:"webhook-url"`
	WebhookPayload string   `mapstructure:"webhook-payload"`
	WebhookHeaders []string `mapstructure:"webhook-headers"`

	// DogStatsdAddr is the address of a dogstatsd instance. If provided,
	// metrics will be sent to that instance
	DogStatsdAddr string `mapstructure:"dog-statsd-addr"`
	// DogStatsdTags are the global tags that should be sent with each packet to dogstatsd
	// It is a list of strings, where each string looks like "my_tag_name:my_tag_value"
	DogStatsdTags []string `mapstructure:"dog-statsd-tags"`
	StatsdAddr    string   `mapstructure:"statsd-addr"`
}

// DefaultBindPort is the default port that dkron will use for Serf communication
const DefaultBindPort int = 8946

func DefaultConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		log.Panic(err)
	}

	tags := map[string]string{"dkron_version": Version}

	return &Config{
		NodeName:          hostname,
		BindAddr:          fmt.Sprintf("0.0.0.0:%d", DefaultBindPort),
		HTTPAddr:          ":8080",
		Discover:          "dkron",
		Backend:           "etcd",
		BackendMachines:   []string{"127.0.0.1:2379"},
		Profile:           "lan",
		Keyspace:          "dkron",
		LogLevel:          "info",
		RPCPort:           6868,
		MailSubjectPrefix: "[Dkron]",
		Tags:              tags,
	}
}

// configFlagSet creates all of our configuration flags.
func ConfigFlagSet() *flag.FlagSet {
	c := DefaultConfig()
	cmdFlags := flag.NewFlagSet("agent flagset", flag.ContinueOnError)

	cmdFlags.Bool("server", false, "start dkron server")
	cmdFlags.String("node-name", c.NodeName, "node name")
	cmdFlags.String("bind-addr", c.BindAddr, "address to bind listeners to")
	cmdFlags.String("advertise-addr", "", "address to advertise to other nodes")
	cmdFlags.String("http-addr", c.HTTPAddr, "HTTP address")
	cmdFlags.String("discover", c.Discover, "mDNS discovery name")
	cmdFlags.String("backend", c.Backend, "store backend")
	cmdFlags.StringSlice("backend-machine", c.BackendMachines, "store backend machines addresses")
	cmdFlags.String("profile", c.Profile, "timing profile to use (lan, wan, local)")
	cmdFlags.StringSlice("join", []string{}, "address of agent to join on startup")
	cmdFlags.StringSlice("tag", []string{}, "tag pair, specified as key=value")
	cmdFlags.String("keyspace", c.Keyspace, "key namespace to use")
	cmdFlags.String("encrypt", "", "encryption key")
	cmdFlags.String("log-level", c.LogLevel, "Log level (debug, info, warn, error, fatal, panic), defaults to info")
	cmdFlags.Int("rpc-port", c.RPCPort, "RPC port")
	cmdFlags.Int("advertise-rpc-port", 0, "advertise RPC port")

	// Notifications
	cmdFlags.String("mail-host", "", "notification mail server host")
	cmdFlags.String("mail-port", "", "port to use for the mail server")
	cmdFlags.String("mail-username", "", "username for the mail server")
	cmdFlags.String("mail-password", "", "password of the mail server")
	cmdFlags.String("mail-from", "", "notification emails from address")
	cmdFlags.String("mail-payload", "", "notification mail payload")
	cmdFlags.String("mail-subject-prefix", c.MailSubjectPrefix, "notification mail subject prefix")

	cmdFlags.String("webhook-url", "", "notification webhook url")
	cmdFlags.String("webhook-payload", "", "notification webhook payload")
	cmdFlags.StringSlice("webhook-header", []string{}, "notification webhook additional header")

	cmdFlags.String("dog-statsd-addr", "", "DataDog Agent address")
	cmdFlags.StringSlice("dog-statsd-tags", []string{}, "Datadog tags, specified as key:value")
	cmdFlags.String("statsd-addr", "", "Statsd Address")

	return cmdFlags
}

// AddrParts returns the parts of the BindAddr that should be
// used to configure Serf.
func (c *Config) AddrParts(address string) (string, int, error) {
	checkAddr := address

START:
	_, _, err := net.SplitHostPort(checkAddr)
	if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
		checkAddr = fmt.Sprintf("%s:%d", checkAddr, DefaultBindPort)
		goto START
	}
	if err != nil {
		return "", 0, err
	}

	// Get the address
	addr, err := net.ResolveTCPAddr("tcp", checkAddr)
	if err != nil {
		return "", 0, err
	}

	return addr.IP.String(), addr.Port, nil
}

// NetworkInterface is used to get the associated network
// interface from the configured value
func (c *Config) NetworkInterface() (*net.Interface, error) {
	if c.Interface == "" {
		return nil, nil
	}
	return net.InterfaceByName(c.Interface)
}

// EncryptBytes returns the encryption key configured.
func (c *Config) EncryptBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.EncryptKey)
}
