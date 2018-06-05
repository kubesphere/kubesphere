// +build !linux,!windows

/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mount

import (
	"errors"
	"os"

	"github.com/golang/glog"
)

type Mounter struct {
	mounterPath string
}

var unsupportedErr = errors.New("util/mount on this platform is not supported")

// New returns a mount.Interface for the current system.
// It provides options to override the default mounter behavior.
// mounterPath allows using an alternative to `/bin/mount` for mounting.
func New(mounterPath string) Interface {
	return &Mounter{
		mounterPath: mounterPath,
	}
}

func (mounter *Mounter) Mount(source string, target string, fstype string, options []string) error {
	return unsupportedErr
}

func (mounter *Mounter) Unmount(target string) error {
	return unsupportedErr
}

// GetMountRefs finds all other references to the device referenced
// by mountPath; returns a list of paths.
func GetMountRefs(mounter Interface, mountPath string) ([]string, error) {
	return []string{}, unsupportedErr
}

func (mounter *Mounter) List() ([]MountPoint, error) {
	return []MountPoint{}, unsupportedErr
}

func (mounter *Mounter) IsMountPointMatch(mp MountPoint, dir string) bool {
	return (mp.Path == dir)
}

func (mounter *Mounter) IsNotMountPoint(dir string) (bool, error) {
	return IsNotMountPoint(mounter, dir)
}

func (mounter *Mounter) IsLikelyNotMountPoint(file string) (bool, error) {
	return true, unsupportedErr
}

func (mounter *Mounter) GetDeviceNameFromMount(mountPath, pluginDir string) (string, error) {
	return "", unsupportedErr
}

func getDeviceNameFromMount(mounter Interface, mountPath, pluginDir string) (string, error) {
	return "", unsupportedErr
}

func (mounter *Mounter) DeviceOpened(pathname string) (bool, error) {
	return false, unsupportedErr
}

func (mounter *Mounter) PathIsDevice(pathname string) (bool, error) {
	return true, unsupportedErr
}

func (mounter *Mounter) MakeRShared(path string) error {
	return unsupportedErr
}

func (mounter *SafeFormatAndMount) formatAndMount(source string, target string, fstype string, options []string) error {
	return mounter.Interface.Mount(source, target, fstype, options)
}

func (mounter *SafeFormatAndMount) diskLooksUnformatted(disk string) (bool, error) {
	return true, unsupportedErr
}

func (mounter *Mounter) GetFileType(pathname string) (FileType, error) {
	return FileType("fake"), unsupportedErr
}

func (mounter *Mounter) MakeDir(pathname string) error {
	return unsupportedErr
}

func (mounter *Mounter) MakeFile(pathname string) error {
	return unsupportedErr
}

func (mounter *Mounter) ExistsPath(pathname string) bool {
	glog.Errorf("%s", unsupportedErr)
	return true
}

func (mounter *Mounter) PrepareSafeSubpath(subPath Subpath) (newHostPath string, cleanupAction func(), err error) {
	return subPath.Path, nil, unsupportedErr
}

func (mounter *Mounter) CleanSubPaths(podDir string, volumeName string) error {
	return unsupportedErr
}

func (mounter *Mounter) SafeMakeDir(pathname string, base string, perm os.FileMode) error {
	return unsupportedErr
}

func (mounter *Mounter) GetMountRefs(pathname string) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (mounter *Mounter) GetFSGroup(pathname string) (int64, error) {
	return -1, errors.New("not implemented")
}
