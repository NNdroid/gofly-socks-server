package reality

import (
	"gofly/pkg/protocol/basic"
	"net"
)

func splitRead(conn net.Conn, expectLen int, packet []byte) (int, error) {
	count := 0
	for {
		err := conn.SetReadDeadline(basic.GetTimeout())
		if err != nil {
			return 0, err
		}
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
