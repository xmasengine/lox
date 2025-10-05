BANK ROM 1024

OPTION EXPLICIT

DEF FN PRINT_XY(X, Y, S) = PRINT AT X+Y*32,S

' Save ram macros
DEF FN SRAM_INIT = POKE $FFFC,$08
DEF FN SRAM_FINI = POKE $FFFC,$00
DEF FN SRAM_SAVE(OFFSET, VALUE) = POKE $8000+OFFSET,VALUE
DEF FN SRAM_READ(OFFSET) = PEEK($8000+OFFSET)

DIM #sram_offset
DIM #sram_data_length
DIM sram_data(64)
DIM #sram_addr
DIM sram_write_error

' volatile __at (0xfffc) unsigned char SRAM_bank_to_be_mapped_on_slot2;
' #define SMS_enableSRAM()        SRAM_bank_to_be_mapped_on_slot2=0x08
' #define SMS_enableSRAMBank(n)   SRAM_bank_to_be_mapped_on_slot2=((((n)<<2)|0x08)&0x0C)
' #define SMS_disableSRAM()       SRAM_bank_to_be_mapped_on_slot2=0x00

' /* SRAM access is as easy as accessing an array of char */
' __at (0x8000) unsigned char SMS_SRAM[];

const SRAM_PORT = $FFFC
const SRAM_OPEN = $08
const SRAM_DONE = $00
CONST SRAM_ADDR = $8000

' ASM sram_init:
' ASM di
' ASM PUSH HL
' ASM LD HL,$08
' ASM LD ($FFFC),HL
' ASM POP HL
' ASM ei
' ASM ret

' ASM sram_fini:
' ASM di
' ASM PUSH HL
' ASM LD HL,$00
' ASM LD ($FFFC),HL
' ASM POP HL
' ASM ei
' ASM ret

main:
	dim arr(5)
	dim I

	SRAM_INIT
	FOR I = 0 TO 4
		arr(I) = SRAM_READ(1+I)
	NEXT I
	SRAM_FINI

	PRINT "-1... "
	FOR I = 0 TO 4
		PRINT CHR$(arr(I))
	NEXT I

	dim res
	SRAM_INIT
	res = SRAM_READ(0)
	SRAM_FINI
	PRINT "0... "
	PRINT CHR$(res)
	PRINT "1... "
	SRAM_INIT
	SRAM_SAVE(0, $76)
	SRAM_FINI
	PRINT "2... "
	SRAM_INIT
	FOR I = 0 TO 4
		SRAM_SAVE(1+I,arr(I))
	NEXT I
	SRAM_FINI
	PRINT "3... "
	SRAM_INIT
	res = SRAM_READ(0)
	SRAM_FINI
	PRINT "4... "
	PRINT CHR$(res)
	if res = $76 then
		PRINT ": WRITE OK"
	else
		PRINT ": WRITE FAIL"
	end if
	SRAM_INIT
	FOR I = 0 TO 4
		arr(I) = SRAM_READ(1+I)
	NEXT I
	SRAM_FINI
	PRINT "5... "
	FOR I = 0 TO 4
		PRINT CHR$(arr(I))
	NEXT I

	while 1
	wend

BANK 50
	BITMAP "01234567"
