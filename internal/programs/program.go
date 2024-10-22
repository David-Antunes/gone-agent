package programs

import "github.com/spf13/viper"

type Program interface {
	SetConfig(viper *viper.Viper)
	Start() error
	Stop() error
	IsRunning() bool
}
