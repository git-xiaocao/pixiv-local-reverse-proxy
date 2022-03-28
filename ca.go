package PixivLocalReverseProxy

import _ "embed"

//go:embed ca/ca.crt
var caCert []byte

//go:embed ca/ca.key
var caKey []byte
