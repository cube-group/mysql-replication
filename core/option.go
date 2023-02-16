package core

type CanalOption struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Destination string
	SoTimOut    int32
	IdleTimeOut int32
	Filter      string
	BatchSize   int32
}
