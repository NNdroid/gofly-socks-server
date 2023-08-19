package reality

import (
	"net"
)

func splitRead(conn net.Conn, expectLen int, packet []byte) (int, error) {
	count := 0
	for {
		n, err := conn.Read(packet[count:])
		if err != nil {
			return count, err
		}
		count += n
		if count == expectLen {
			break
		}
	}
	return count, nil
}
