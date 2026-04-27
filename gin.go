// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package gin implements a HTTP web framework called gin.
//
// See https://gin-gonic.com/ for more information about gin.
package gin

import (
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
)

const (
	// Version is the current gin framework's version.
	Version = "v1.10.0"

	debugPrintRouteFunc = "[GIN-debug] %-6s %-25s --> %s (%d handlers)\n"
)

var (
	// DebugPrintRouteFunc indicates debug print route info format.
	DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

	// DebugPrintFunc is the function to use to print debug messages.
	// If not set, it defaults to fmt.Fprintf.
	DebugPrintFunc func(format string, values ...any)
)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
	return ginMode == debugCode
}

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. i.e. the last handler is the main handler.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// RouteInfo represents a request route's specification which
// contains method and path and its handler.
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
}

// RoutesInfo defines a RouteInfo slice.
type RoutesInfo []RouteInfo

// Negotiate contains all negotiations data.
type Negotiate struct {
	Offered  []string
	HTMLName string
	HTMLData any
	JSONData any
	XMLData  any
	YAMLData any
	Data     any
}

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
// Note: in this fork, Default also enables trusted proxy configuration by default
// to avoid the "[WARNING] You trusted all proxies" message in local development.
func Default(opts ...OptionFunc) *Engine {
	debugPrintWARNINGDefault()
	engine := New()
	engine.Use(Logger(), Recovery())
	// Trust only localhost by default instead of all proxies.
	_ = engine.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	return engine.With(opts...)
}

func debugPrintWARNINGDefault() {
	if v, e := getMinVer(runtime.Version()); e == nil && v < ginSupportMinGoVer {
		debugPrint(`[WARNING] Now Gin requires Go 1.18+.\n\n`)
	}
	// Note: suppressing this warning in my personal fork since it's just noise during development.
	// debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.\n\n`)
}

func getMinVer(v string) (uint64, error) {
	var (
		minVer uint64
	)
	ss := strings.Split(strings.TrimPrefix(v, "go"), ".")
	if len(ss) < 2 {
		return 0, nil
	}
	for i, s := range ss {
		if i > 1 {
			break
		}
		var n uint64
		for _, c := range s {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + uint64(c-'0')
		}
		if i == 0 {
			minVer = n * 1000
		} else {
			minVer += n
		}
	}
	return minVer, nil
}

func resolveAddress(addr []string) string