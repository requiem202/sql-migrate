package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/requiem202/sql-migrate"
	"gopkg.in/gorp.v1"
	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"github.com/spf13/viper"
)

var dialects = map[string]gorp.Dialect{
	"sqlite3":  gorp.SqliteDialect{},
	"postgres": gorp.PostgresDialect{},
	"mysql":    gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"},
}

var ConfigFile string
var ConfigEnvironment string

func ConfigFlags(f *flag.FlagSet) {
	f.StringVar(&ConfigFile, "config", "dbconfig.yml", "Configuration file to use.")
	//f.StringVar(&ConfigFile, "config", "dbconfig.toml", "Configuration file to use.")
	f.StringVar(&ConfigEnvironment, "env", "development", "Environment to use.")
}

type Environment struct {
	Dialect    string `yaml:"dialect"`
	DataSource string `yaml:"datasource"`
	Dir        string `yaml:"dir"`
	TableName  string `yaml:"table"`
	SchemaName string `yaml:"schema"`
}

func ReadConfig() (map[string]*Environment, error) {
	if strings.HasSuffix(ConfigFile, ".yaml", ) ||strings.HasSuffix(ConfigFile, ".yml") {
		file, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			return nil, err
		}

		config := make(map[string]*Environment)
		err = yaml.Unmarshal(file, config)
		if err != nil {
			return nil, err
		}

		return config, nil
	} else {
		return ReadConfigToml()
	}
}

func ReadConfigToml() (map[string]*Environment, error) {
	viper.SetConfigFile(ConfigFile)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := viper.GetStringMap("database")
	config := make(map[string]*Environment)

	for key, m := range cfg {
		dict := m.(map[string]interface{})
		env := &Environment{}
		if _, ok := dict["dialect"]; ok {
			env.Dialect = dict["dialect"].(string)
		}
		if _, ok := dict["datasource"]; ok {
			env.DataSource = dict["datasource"].(string)
		}
		if _, ok := dict["dir"]; ok {
			env.Dir = dict["dir"].(string)
		}
		if _, ok := dict["table"]; ok {
			env.TableName = dict["table"].(string)
		}
		if _, ok := dict["schema"]; ok {
			env.SchemaName = dict["schema"].(string)
		}
		config[key] = env
	}

	return config, nil
}

func appendEnvToFileName(fileName string) string {
	ss := strings.Split(ConfigFile, ".")
	return strings.Join(append(ss[:len(ss)-1], ConfigEnvironment, ss[len(ss)]), ".")
}

func GetEnvironment() (*Environment, error) {
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {

		ConfigFile = "dbconfig.toml"
		// try reading a dbconfig.environment.toml first
		envConfigName := appendEnvToFileName(ConfigFile)
		if _, err := os.Stat(envConfigName); err == nil {
			ConfigFile = envConfigName
		}
	}
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	env := config[ConfigEnvironment]
	if env == nil {
		return nil, errors.New("No environment: " + ConfigEnvironment)
	}

	if env.Dialect == "" {
		return nil, errors.New("No dialect specified")
	}

	if env.DataSource == "" {
		return nil, errors.New("No data source specified")
	}
	env.DataSource = os.ExpandEnv(env.DataSource)

	if env.Dir == "" {
		env.Dir = "migrations"
	}

	if env.TableName != "" {
		migrate.SetTable(env.TableName)
	}

	if env.SchemaName != "" {
		migrate.SetSchema(env.SchemaName)
	}

	return env, nil
}

func GetConnection(env *Environment) (*sql.DB, string, error) {
	db, err := sql.Open(env.Dialect, env.DataSource)
	if err != nil {
		return nil, "", fmt.Errorf("Cannot connect to database: %s", err)
	}

	// Make sure we only accept dialects that were compiled in.
	_, exists := dialects[env.Dialect]
	if !exists {
		return nil, "", fmt.Errorf("Unsupported dialect: %s", env.Dialect)
	}

	return db, env.Dialect, nil
}
