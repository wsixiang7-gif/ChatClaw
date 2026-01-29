# WillChat

## 开发

```bash
wails3 dev
```

## 构建开发环境客户端（development）

构建 development 版本：

```bash
wails3 build DEV=true
```

## 构建生产环境客户端（production）

构建 production 版本：

```bash
wails3 build
```

打包：

```bash
wails3 package
```

## Windows 多架构打包（production）

```bash
# amd64
wails3 task windows:build ARCH=amd64
wails3 task windows:package ARCH=amd64

# arm64
wails3 task windows:build ARCH=arm64
wails3 task windows:package ARCH=arm64
```

## macOS 多架构打包（production）

```bash
# arm64 / amd64
wails3 task package ARCH=arm64
wails3 task package ARCH=amd64

# universal（二进制 + .app）
wails3 task darwin:package:universal
```

## macOS 生成 DMG（production）

```bash
wails3 task darwin:create:dmg ARCH=arm64
wails3 task darwin:create:dmg ARCH=amd64
wails3 task darwin:create:dmg UNIVERSAL=true
```
