# SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
# SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
# 
# SPDX-License-Identifier: GPL-3.0-only

@default:
	just --list

update-parser:
	kaitai-struct-compiler --target go --go-package parser --outdir . kaitai/*.ksy

build:
	go build -o sbt-inf .

build-win:
	GOOS=windows GOARCH=amd64 go build -o sbt-inf.exe .

generate-mos: build
	./sbt-inf update-bam inputs/1-startbut/startbut.toml
	./sbt-inf update-bam inputs/2-cgattr/cgattr.toml 

run *params: build
	./sbt-inf {{params}}
