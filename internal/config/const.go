package config

type AppMode string

var AppModeType = struct {
	StartUI     AppMode
	DisplayInfo AppMode
	HandleParam AppMode
}{
	StartUI:     "START_UI",
	DisplayInfo: "DISPLAY_INFO",
	HandleParam: "HANDLE_PARAM",
}
