package config

type Config struct {
	DB struct {
		Host   string
		Port   int
		User   string
		Pass   string
		DBName string
	}
	Mailer struct {
		Host string
		Port int
		User string
		Pass string
	}
	Server struct {
		Addr string
	}
}

func Default() *Config {
	return &defaultConfig
}

var defaultConfig = Config{
	DB: struct {
		Host   string
		Port   int
		User   string
		Pass   string
		DBName string
	}{
		Host:   "127.0.0.1",
		Port:   3306,
		User:   "root",
		Pass:   "12345678",
		DBName: "ServiceA",
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
		Addr: "0.0.0.0:10001",
	},
}
