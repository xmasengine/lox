.POSIX:
.SUFFIXES:
#
# Makefile for Lord Of Xmas.
#
BFLAGS = --sms
BC = cvbasic
BLIB = $(HOME)/opt/nanochess/CVBasic
ASM = gasm80
ASMFLAGS = -sms
EMU=mg -tv -scale 5
EMU2=gearsystem

MAPS=./map/m0003-church.xml.bas
MUSIC=./music/kirie.bas ./music/gloria.bas ./music/nocturne.bas
IMAGE=img/nun.bas
DEPS=nun.bas sprite1.bas tile2.bas map1.bas $(MAPS) $(MUSIC) $(IMAGE)

build: build/lox.sms
	@echo "Built $<"

compile: build/lox.asm
	@echo "Compiled $<"

build/lox.sms: build/lox.asm
	$(ASM) $< -o $@ $(ASMFLAGS)

build/lox.asm: lox.bas $(DEPS)
	@$(BC) $(BFLAGS) $< $@ $(BLIB) > /dev/null

run: build/lox.sms
	@$(EMU) build/lox.sms

gs: build/lox.sms
	@$(EMU2) build/lox.sms

clean:
	@rm build/lox.asm build/lox.sms

img/nun.bas: img/nun.png
	tmscolor -sms -p2 -b -t128 img/nun.png img/nun.bas nun

.PHONY: run clean build gs
