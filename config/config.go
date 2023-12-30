package config

type Schedule struct {
	Groups []struct {
		ID   int    `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"groups"`

	Teachers []struct {
		Name string `yaml:"name"`
	} `yaml:"teachers"`
}

type Telegram struct {
	Token   string `yaml:"token"`
	Support struct {
		ID   int64  `yaml:"id"`
		Href string `yaml:"href"`
	} `yaml:"support"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Pass string `yaml:"password"`
}
