package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttBroker = "tcp://broker.com:1883" // ganti IP & port broker
	mqttUser   = "user"                      // ganti user
	mqttPass   = "pass"                      // ganti password
	mqttTopic  = "laptop/username/battery"
	interval   = 5 * time.Minute // publish setiap 5 menit
)

func getBatteryLevel() (float64, error) {
	cmd := `upower -i $(upower -e | grep BAT) | grep percentage | awk '{print $2}'`
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return 0, fmt.Errorf("gagal ambil data batre: %v", err)
	}

	str := strings.TrimSpace(string(out))
	str = strings.TrimSuffix(str, "%")
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("gagal parse angka: %v", err)
	}
	return val, nil
}

func main() {
	opts := mqtt.NewClientOptions().
		AddBroker(mqttBroker).
		SetUsername(mqttUser).
		SetPassword(mqttPass)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("[INFO] MQTT Connected ✅")

	for {
		level, err := getBatteryLevel()
		if err != nil {
			fmt.Println("[WARN]", err)
			time.Sleep(interval)
			continue
		}

		payload := fmt.Sprintf("%.1f", level)
		token := client.Publish(mqttTopic, 0, false, payload)
		token.Wait()
		if token.Error() != nil {
			fmt.Println("[ERR] Gagal publish:", token.Error())
		} else {
			fmt.Printf("[OK] Battery: %.1f%%\n", level)
		}

		time.Sleep(interval)
	}
}
