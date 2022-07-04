package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashcloak/Meson/client/config"
	"github.com/katzenpost/client/utils"
)

func register() (*Client, *Session, *utils.ServiceDescriptor) {
	cfg, err := config.LoadFile("client_test.toml")
	if err != nil {
		panic(err)
	}
	_ = cfg.UpdateTrust()

	_ = cfg.SaveConfig("client_test.toml")

	linkKey := AutoRegisterRandomClient(cfg)

	c, err := NewFromConfig(cfg, "echo")
	if err != nil {
		c.Shutdown()
		panic(err)
	}

	s, err := c.NewSession(linkKey)
	if err != nil {
		c.Shutdown()
		panic(err)
	}

	serviceDesc, err := s.GetService("echo")
	if err != nil {
		c.Shutdown()
		panic(err)
	}

	return c, s, serviceDesc
}

func TestBasicBlockingSend(t *testing.T) {
	c, s, serviceDesc := register()

	fmt.Printf("Sending Sphinx packet payload to: %s@%s\n", serviceDesc.Name, serviceDesc.Provider)
	resp, err := s.BlockingSendUnreliableMessage(serviceDesc.Name, serviceDesc.Provider, []byte(`Data encryption is used widely today!`))
	if err != nil {
		c.Shutdown()
		panic(err)
	}
	payload, err := ValidateReply(resp)
	if err != nil {
		c.Shutdown()
		panic(err)
	}
	fmt.Printf("Return: %s\n", payload)

	c.Shutdown()
}

func TestBasicNonBlockingSend(t *testing.T) {
	c, s, serviceDesc := register()

	fmt.Printf("Sending Sphinx packet payload to: %s@%s\n", serviceDesc.Name, serviceDesc.Provider)
	id, err := s.SendUnreliableMessage(serviceDesc.Name, serviceDesc.Provider, []byte(`Data encryption is used widely today!`))
	if err != nil {
		c.Shutdown()
		panic(err)
	}
	fmt.Printf("Return: %s\n", id)

	c.Shutdown()
}

func TestSendingBlockingConcurrently(t *testing.T) {
	c, s, serviceDesc := register()
	done := make(chan struct{})
	timeout := time.After(time.Second * 80)

	fmt.Printf("Concurrently Sending 20 Sphinx packet payloads to: %s@%s\n", serviceDesc.Name, serviceDesc.Provider)
	for i := 0; i < 20; i++ {
		go func(i int) {
			fmt.Printf("Iteration %d\n", i+1)
			resp, err := s.BlockingSendUnreliableMessage(serviceDesc.Name, serviceDesc.Provider, []byte(fmt.Sprintf("Data encryption is used widely today! Iteration: %d", i+1)))
			if err != nil {
				c.Shutdown()
				panic(err)
			}

			payload, err := ValidateReply(resp)
			if err != nil {
				c.Shutdown()
				panic(err)
			}

			fmt.Printf("Return: %s\n", payload)
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-timeout:
			c.Shutdown()
			panic("Timeout")
		}
	}

	c.Shutdown()
}

func TestSendingNonBlockingConcurrently(t *testing.T) {
	c, s, serviceDesc := register()
	done := make(chan struct{})
	timeout := time.After(time.Second * 80)

	fmt.Printf("Concurrently Sending 20 Sphinx packet payloads to: %s@%s\n", serviceDesc.Name, serviceDesc.Provider)
	for i := 0; i < 20; i++ {
		go func(i int) {
			fmt.Printf("Iteration %d\n", i+1)
			id, err := s.SendUnreliableMessage(serviceDesc.Name, serviceDesc.Provider, []byte(fmt.Sprintf("Data encryption is used widely today! Iteration: %d", i+1)))
			if err != nil {
				c.Shutdown()
				panic(err)
			}

			fmt.Printf("Msg ID: %s\n", id)
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-timeout:
			c.Shutdown()
			panic("Timeout")
		}
	}

	c.Shutdown()
}

func TestSendingDropLoopDecoy(t *testing.T) {
	c, s, serviceDesc := register()
	done := make(chan struct{})
	timeout := time.After(time.Second * 80)

	fmt.Printf("Sending 20 Drop and Loop decoys to: %s@%s\n", serviceDesc.Name, serviceDesc.Provider)
	for i := 0; i < 20; i++ {
		go func(i int) {
			fmt.Printf("Iteration %d\n", i+1)
			s.sendDropDecoy()
			s.sendLoopDecoy()

			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 20; i++ {
		select {
		case <-done:
		case <-timeout:
			c.Shutdown()
			panic("Timeout")
		case <-s.fatalErrCh:
			c.Shutdown()
			panic("Fatal error")
		}
	}

	c.Shutdown()
}
