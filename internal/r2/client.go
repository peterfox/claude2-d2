package r2

import (
	"errors"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

var (
	connectionServiceUUID = bluetooth.NewUUID([16]byte{0x00, 0x02, 0x00, 0x01, 0x57, 0x4F, 0x4F, 0x20, 0x53, 0x70, 0x68, 0x65, 0x72, 0x6F, 0x21, 0x21})
	handleCharUUID        = bluetooth.NewUUID([16]byte{0x00, 0x02, 0x00, 0x02, 0x57, 0x4F, 0x4F, 0x20, 0x53, 0x70, 0x68, 0x65, 0x72, 0x6F, 0x21, 0x21})
	connectCharUUID       = bluetooth.NewUUID([16]byte{0x00, 0x02, 0x00, 0x05, 0x57, 0x4F, 0x4F, 0x20, 0x53, 0x70, 0x68, 0x65, 0x72, 0x6F, 0x21, 0x21})
	mainServiceUUID       = bluetooth.NewUUID([16]byte{0x00, 0x01, 0x00, 0x01, 0x57, 0x4F, 0x4F, 0x20, 0x53, 0x70, 0x68, 0x65, 0x72, 0x6F, 0x21, 0x21})
	mainCharUUID          = bluetooth.NewUUID([16]byte{0x00, 0x01, 0x00, 0x02, 0x57, 0x4F, 0x4F, 0x20, 0x53, 0x70, 0x68, 0x65, 0x72, 0x6F, 0x21, 0x21})
)

type Client struct {
	device   bluetooth.Device
	mainChar bluetooth.DeviceCharacteristic
}

func isR2D2(name string) bool {
	return len(name) >= 3 && name[:3] == "D2-"
}

// FindDevice scans for an R2-D2 and returns the first result found.
func FindDevice(adapter *bluetooth.Adapter) (address string, name string, err error) {
	foundCh := make(chan bluetooth.ScanResult, 1)

	scanErr := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if isR2D2(result.LocalName()) || result.AdvertisementPayload.HasServiceUUID(connectionServiceUUID) {
			foundCh <- result
			_ = adapter.StopScan()
		}
	})
	if scanErr != nil {
		return "", "", fmt.Errorf("scan failed: %w", scanErr)
	}

	select {
	case result := <-foundCh:
		return result.Address.String(), result.LocalName(), nil
	case <-time.After(15 * time.Second):
		return "", "", errors.New("timed out scanning — is R2-D2 powered on?")
	}
}

// Connect scans for an R2-D2 and connects to the first one found.
func Connect(adapter *bluetooth.Adapter) (*Client, string, error) {
	address, name, err := FindDevice(adapter)
	if err != nil {
		return nil, "", err
	}
	client, err := ConnectByAddress(adapter, address)
	return client, name, err
}

// ConnectByAddress connects directly to a known BLE address, skipping the scan.
func ConnectByAddress(adapter *bluetooth.Adapter, address string) (*Client, error) {
	var bleAddr bluetooth.Address
	bleAddr.Set(address)

	device, err := adapter.Connect(bleAddr, bluetooth.ConnectionParams{})
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	return establishSession(device)
}

func establishSession(device bluetooth.Device) (*Client, error) {
	connServices, err := device.DiscoverServices([]bluetooth.UUID{connectionServiceUUID})
	if err != nil || len(connServices) == 0 {
		return nil, fmt.Errorf("discover connection service failed: %w", err)
	}

	connChars, err := connServices[0].DiscoverCharacteristics([]bluetooth.UUID{handleCharUUID, connectCharUUID})
	if err != nil || len(connChars) < 2 {
		return nil, fmt.Errorf("discover connection characteristics failed: %w", err)
	}

	var handleChar, connectChar bluetooth.DeviceCharacteristic
	for _, ch := range connChars {
		switch ch.UUID() {
		case handleCharUUID:
			handleChar = ch
		case connectCharUUID:
			connectChar = ch
		}
	}

	if err := handleChar.EnableNotifications(func(buf []byte) {}); err != nil {
		return nil, fmt.Errorf("enable handle notifications failed: %w", err)
	}

	if _, err := connectChar.WriteWithoutResponse(handshakeBytes); err != nil {
		return nil, fmt.Errorf("handshake write failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	mainServices, err := device.DiscoverServices([]bluetooth.UUID{mainServiceUUID})
	if err != nil || len(mainServices) == 0 {
		return nil, fmt.Errorf("discover main service failed: %w", err)
	}

	mainChars, err := mainServices[0].DiscoverCharacteristics([]bluetooth.UUID{mainCharUUID})
	if err != nil || len(mainChars) == 0 {
		return nil, fmt.Errorf("discover main characteristic failed: %w", err)
	}

	client := &Client{
		device:   device,
		mainChar: mainChars[0],
	}

	if err := client.mainChar.EnableNotifications(func(buf []byte) {}); err != nil {
		return nil, fmt.Errorf("enable main notifications failed: %w", err)
	}

	if err := client.sendInit(); err != nil {
		return nil, fmt.Errorf("init failed: %w", err)
	}

	time.Sleep(5 * time.Second)

	return client, nil
}

func (c *Client) sendInit() error {
	_, err := c.mainChar.WriteWithoutResponse(initPacket())
	return err
}

func (c *Client) Animate(animationID byte) error {
	_, err := c.mainChar.WriteWithoutResponse(animatePacket(animationID))
	return err
}

func (c *Client) StopAnimation() error {
	_, err := c.mainChar.WriteWithoutResponse(stopAnimationPacket())
	return err
}

func (c *Client) Disconnect() error {
	return c.device.Disconnect()
}
