# Makefile for Lord Of Xmas.
#
BFLAGS = --sms
BC = cvbasic
BLIB = $(HOME)/opt/nanochess/CVBasic
ASM = gasm80
ASMFLAGS = -sms
EMU=mg -scale 5
DEPS=nun.bas sprite1.bas tile2.bas map1.bas

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

clean:
	@rm build/lox.asm build/lox.sms

.PHONY: run clean build
