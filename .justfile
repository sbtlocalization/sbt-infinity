# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
# 
# SPDX-License-Identifier: GPL-3.0-only

update-parser:
	kaitai-struct-compiler --target go --go-package parser --outdir . kaitai/bam.ksy

build:
	go build

build-win:
	GOOS=windows GOARCH=amd64 go build -o infinity-tools.exe .

generate-mos: build
	./infinity-tools update-bam inputs/1-startbut/startbut.toml
	./infinity-tools update-bam inputs/2-cgattr/cgattr.toml 
