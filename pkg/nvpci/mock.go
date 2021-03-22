/*
 * Copyright (c) 2021, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nvpci

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/nvidia/cloud-native/go-nvlib/pkg/nvpci/bytes"
)

type MockNvpci struct {
	*nvpci
}

var _ Interface = (*MockNvpci)(nil)

func NewMockNvpci() (mock *MockNvpci, rerr error) {
	rootDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr != nil {
			os.RemoveAll(rootDir)
		}
	}()

	mock = &MockNvpci{
		&nvpci{rootDir},
	}

	return mock, nil
}

func (m *MockNvpci) Cleanup() {
	os.RemoveAll(m.pciDevicesRoot)
}

func (m *MockNvpci) AddMockA100(address string) error {
	deviceDir := filepath.Join(m.pciDevicesRoot, address)
	err := os.MkdirAll(deviceDir, 0755)
	if err != nil {
		return err
	}

	vendor, err := os.Create(filepath.Join(deviceDir, "vendor"))
	if err != nil {
		return err
	}
	_, err = vendor.WriteString(fmt.Sprintf("0x%x", pciNvidiaVendorID))
	if err != nil {
		return err
	}

	class, err := os.Create(filepath.Join(deviceDir, "class"))
	if err != nil {
		return err
	}
	_, err = class.WriteString(fmt.Sprintf("0x%x", pci3dControllerClass))
	if err != nil {
		return err
	}

	device, err := os.Create(filepath.Join(deviceDir, "device"))
	if err != nil {
		return err
	}
	_, err = device.WriteString("0x20bf")
	if err != nil {
		return err
	}

	config, err := os.Create(filepath.Join(deviceDir, "config"))
	if err != nil {
		return err
	}
	_data := make([]byte, pciCfgSpaceStandardSize)
	data := bytes.New(&_data)
	data.Write16(0, pciNvidiaVendorID)
	data.Write16(2, uint16(0x20bf))
	_, err = config.Write(*data.Raw())
	if err != nil {
		return err
	}

	bar0 := []uint64{0x00000000c2000000, 0x00000000c2ffffff, 0x0000000000040200}
	resource, err := os.Create(filepath.Join(deviceDir, "resource"))
	_, err = resource.WriteString(fmt.Sprintf("0x%x 0x%x 0x%x", bar0[0], bar0[1], bar0[2]))
	if err != nil {
		return err
	}

	pmcID := uint32(0x170000a1)
	resource0, err := os.Create(filepath.Join(deviceDir, "resource0"))
	if err != nil {
		return err
	}
	_data = make([]byte, bar0[1]-bar0[0]+1)
	data = bytes.New(&_data).LittleEndian()
	data.Write32(0, pmcID)
	_, err = resource0.Write(*data.Raw())
	if err != nil {
		return err
	}

	return nil
}
