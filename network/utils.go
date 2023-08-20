package network

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	nm "github.com/Wifx/gonetworkmanager"
	"github.com/darkhz/bluetuith/cmd"
)

// isDeviceAddrExist checks if the device's address is present
// in the connection's settings.
func isDeviceAddrExist(conn nm.Connection, connType, bdaddr string) (bool, error) {
	settings, err := conn.GetSettings()
	if err != nil {
		return false, err
	}

	addr, ok := settings["bluetooth"]["bdaddr"].([]byte)
	if ok {
		bdtype, ok := settings["bluetooth"]["type"]
		if ok && getMacAddress(addr) == bdaddr && bdtype == connType {
			return true, nil
		}
	}

	return false, nil
}

// checkSettings checks and modifies the device connection's settings.
func checkSettings(conn nm.Connection, connType string) error {
	if connType != "dun" {
		return nil
	}

	settings, err := conn.GetSettings()
	if err != nil {
		return err
	}

	gsmSettings, ok := settings["gsm"]
	if !ok {
		return NMSettingModifyError
	}

	apn := cmd.GetProperty("gsm-apn")
	number := cmd.GetProperty("gsm-number")

	if apn != gsmSettings["apn"] {
		gsmSettings["apn"] = apn
	}

	if number != gsmSettings["number"] {
		gsmSettings["number"] = number
	}

	delete(settings, "ipv6")

	return conn.Update(settings)
}

// getSettings returns a connection setting.
func getSettings(bdaddr []byte, name, connType, uuid string) map[string]map[string]interface{} {
	name = fmt.Sprintf("%s Access Point (%s)",
		name, strings.ToUpper(connType),
	)

	settings := map[string]map[string]interface{}{
		"connection": {
			"id":          name,
			"type":        "bluetooth",
			"uuid":        uuid,
			"autoconnect": false,
		},
		"bluetooth": {
			"bdaddr": bdaddr,
			"type":   connType,
		},
	}

	if connType == "dun" {
		apn := cmd.GetProperty("gsm-apn")
		number := cmd.GetProperty("gsm-number")

		settings["gsm"] = map[string]interface{}{
			"apn":    apn,
			"number": number,
		}
	}

	return settings
}

// getMacAddress gets a MAC address from a bluetooth address byte array.
func getMacAddress(addr []byte) string {
	var macAddr []string

	for _, addrByte := range addr {
		macAddr = append(macAddr, hex.EncodeToString([]byte{addrByte}))
	}

	return strings.ToUpper(strings.Join(macAddr, ":"))
}

// getBDAddress gets a bluetooth address byte array from a MAC address.
func getBDAddress(addr string) ([]byte, error) {
	var bdAddr []byte

	for _, addrElem := range strings.Split(strings.ToLower(addr), ":") {
		addrByteElem, err := strconv.ParseUint(addrElem, 16, 8)
		if err != nil {
			return bdAddr, err
		}

		bdAddr = append(bdAddr, byte(addrByteElem))
	}

	return bdAddr, nil
}
