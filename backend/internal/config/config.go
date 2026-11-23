package config

import (
	"diploma/internal/scope"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"strings"
)

type Config struct{ DSN, Token, Actor, Origin, Listen string }

func Load() (Config, error) {
	c := Config{DSN: os.Getenv("DATABASE_URL"), Token: os.Getenv("APP_API_TOKEN"), Actor: os.Getenv("APP_OPERATOR_ID"), Origin: os.Getenv("JUICE_SHOP_ORIGIN"), Listen: os.Getenv("APP_LISTEN")}
	b, e := hex.DecodeString(c.Token)
	if e != nil || len(b) < 32 {
		return c, errors.New("APP_API_TOKEN must be at least 32 random bytes encoded as hex")
	}
	if c.DSN == "" || c.Actor == "" || len(c.Actor) > 200 {
		return c, errors.New("DATABASE_URL and APP_OPERATOR_ID are required")
	}
	if c.Origin == "" {
		c.Origin = "http://juice-shop:3000"
	}
	l, e := scope.Parse(c.Origin)
	if e != nil || l.Path != "/" || len(l.QueryNames) != 0 {
		return c, errors.New("JUICE_SHOP_ORIGIN must be a valid origin")
	}
	c.Origin = l.Origin
	if c.Listen == "" {
		c.Listen = "127.0.0.1:8080"
	}
	h, _, e := net.SplitHostPort(c.Listen)
	if e != nil || !(h == "127.0.0.1" || h == "::1" || h == "0.0.0.0" && os.Getenv("APP_CONTAINER") == "true") {
		return c, errors.New("APP_LISTEN must bind loopback (container opt-in required for wildcard)")
	}
	if strings.ContainsAny(c.Actor, "\r\n") {
		return c, errors.New("Invalid operator identifier")
	}
	return c, nil
}
