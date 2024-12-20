package config

import "github.com/yyboo586/common/dbUtils"

type Config struct {
	DBConfig dbUtils.Config
	Mailer   struct {
		Host string
		Port int
		User string
		Pass string
	}
	Server struct {
		Addr string
	}
	Logger struct {
		Level string
	}
}

func Default() *Config {
	return &defaultConfig
}

var defaultConfig = Config{
	DBConfig: dbUtils.Config{
		User:   "root",
		Passwd: "12345678",
		Host:   "127.0.0.1",
		Port:   3306,
		DBName: "IAMService",
	},
	Mailer: struct {
		Host string
		Port int
		User string
		Pass string
	}{
		Host: "sandbox.smtp.mailtrap.io",
		Port: 25,
		User: "8df5de08b5f13f",
		Pass: "fcb5034938135d",
	},
	Server: struct {
		Addr string
	}{
		Addr: "127.0.0.1:12000",
	},
	Logger: struct {
		Level string
	}{
		Level: "debug",
	},
}
