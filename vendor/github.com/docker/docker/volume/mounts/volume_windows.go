package mounts // import "github.com/docker/docker/volume/mounts"

func (p *windowsParser) HasResource(m *MountPoint, absolutePath string) bool {
	return false
}
func (p *linuxParser) HasResource(m *MountPoint, absolutePath string) bool {
	return false
}
