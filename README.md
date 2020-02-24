# Get Running Configuration from Cisco IOS

This code works at best effort; therefore, the SSH client tries to connect to servers even **with not recommended Kex and Ciphers**:

```golang
DIAL:
	conn, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		m := regexp.MustCompile(`server offered: \[(.*)\]`)

		errString := err.Error()
		if strings.Contains(errString, "key exchange") {
			sshConfig.Config.KeyExchanges = strings.Split(m.FindStringSubmatch(errString)[1], " ")
			goto DIAL
		}
		if strings.Contains(errString, "server cipher") {
			sshConfig.Config.Ciphers = strings.Split(m.FindStringSubmatch(errString)[1], " ")
			goto DIAL
		}
		return nil, err
	}

```

`goto` proves very useful in this case because it reduced the amount of code to handle all these exceptions.

## Usage:

```
./getrun -t {target} -P {port} -u {username} -p {password}
```
