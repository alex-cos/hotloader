# hotloader

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![Test Status](https://github.com/alex-cos/hotloader/actions/workflows/test.yml/badge.svg)](https://github.com/alex-cos/hotloader/actions/workflows/test.yml)
[![Lint Status](https://github.com/alex-cos/hotloader/actions/workflows/lint.yml/badge.svg)](https://github.com/alex-cos/hotloader/actions/workflows/lint.yml)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex-cos/hotloader)](https://goreportcard.com/report/github.com/alex-cos/hotloader)

A Go library for hot-reloading configuration files at runtime via OS signals.

## Overview

`hotloader` watches for OS signals (e.g., `SIGUSR1`) and automatically re-reloads a configuration file into memory when received. This enables live configuration changes without restarting the application.

## Installation

```bash
go get github.com/alex-cos/hotloader
```

## Usage

### 1. Implement the `Loader` interface

Your configuration struct must implement the `Loader` interface:

```go
type Loader interface {
    Load(filename string) error
    InitDefault()
}
```

Example:

```go
type Config struct {
    Port    int
    Debug   bool
}

func (c *Config) Load(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, c)
}

func (c *Config) InitDefault() {
    c.Port = 8080
    c.Debug = false
}
```

### 2. Create a HotLoader

```go
config := &Config{}
hl := hotloader.New("config.json", config, syscall.SIGUSR1)
defer hl.Stop()
```

### 3. Access the configuration

```go
cfg := hl.Get().(*Config)
fmt.Println(cfg.Port)
```

### 4. Set callbacks (optional)

```go
hl.SetBeforeFunc(func(l hotloader.Loader) {
    log.Println("Reloading configuration...")
})

hl.SetAfterFunc(func(l hotloader.Loader) {
    log.Println("Configuration reloaded successfully")
})

hl.SetErrorFunc(func(e error) {
    log.Printf("Failed to reload configuration: %v", e)
})
```

## API

### `New(filename string, loader Loader, signals ...os.Signal) HotLoader`

Creates a new HotLoader that watches the given file and listens for the specified OS signals.

### `HotLoader` interface

| Method | Description |
| -------- | ------------- |
| `Load() error` | Manually trigger a configuration reload |
| `Get() Loader` | Get the current loaded configuration (thread-safe) |
| `SetBeforeFunc(f func(l Loader))` | Set callback before reload |
| `SetAfterFunc(f func(l Loader))` | Set callback after successful reload |
| `SetErrorFunc(f func(e error))` | Set callback on reload error |

### `Loader` interface

| Method | Description |
| -------- | ------------- |
| `Load(filename string) error` | Parse the configuration file |
| `InitDefault()` | Initialize default values |

## Signal Handling

On Unix-like systems, send `SIGUSR1` to trigger a reload:

```bash
kill -USR1 <pid>
```

On Windows, use `Load()` manually as OS signals are limited.

## Thread Safety

The `Get()` method is safe for concurrent use. Configuration is protected by a read-write mutex.

## License

MIT
