package PixivLocalReverseProxy

import _ "embed"

//go:embed ca/ca.pem
var caCert []byte

//go:embed ca/ca.key.pem
var caKey []byte
