//go:build windows

package tray

func setupTrayImpl(showWindow func(), itemClick func(int), clear func(), quit func()) {}
func updateMenuImpl(items []string) {}
