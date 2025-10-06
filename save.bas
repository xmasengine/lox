BANK ROM 1024

OPTION EXPLICIT

DEF FN PRINT_XY(X, Y, S) = PRINT AT X+Y*32,S

' Save ram macros
const #SRAM_PORT = $FFFC
CONST #SRAM_ADDR = $8000
DEF FN SRAM_INIT = POKE #SRAM_PORT,$08
DEF FN SRAM_FINI = POKE #SRAM_PORT,$00
DEF FN SRAM_SAVE(OFFSET, VALUE) = POKE #SRAM_ADDR+OFFSET,VALUE
DEF FN SRAM_READ(OFFSET) = PEEK(#SRAM_ADDR+OFFSET)

DEF FN SRAM_READA(A,L,I,O) = FOR I = 0 TO (L-1): A(I) = SRAM_READ(O+I): NEXT I
DEF FN SRAM_SAVEA(A,L,I,O) = FOR I = 0 TO (L-1): SRAM_SAVE(O+I,A(I)): NEXT I

DEF FN SRAM_READP(P,L,I,O) = FOR I = 0 TO (L-1): P(I) = SRAM_READ(P+I): NEXT I
DEF FN SRAM_SAVEP(P,L,I,O) = FOR I = 0 TO (L-1): SRAM_SAVE(P+I,P(I)): NEXT I

DEF FN PRINTA(A,L, I) = FOR I = 0 TO (L-1): PRINT CHR$(arr(I)): NEXT I

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
	SRAM_READA(arr, 5, I, 1)
	SRAM_FINI

	PRINT "-1... "
	PRINTA(arr, 5, I)

	FOR I = 0 TO 5: arr(I) = I+70: NEXT I
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
	SRAM_SAVEA(arr, 5, I, 1)
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
	dim try
	SRAM_INIT
	SRAM_READA(arr, 5, I, 1)
	sram_label: DATA BYTE #SRAM_ADDR
	try = sram_label(3)
	SRAM_FINI
	PRINT "5... "
	PRINTA(arr, 5, I)
	PRINT "6... "
	PRINT CHR$(try)
	PRINT "7... "

	dim #foo
	#foo = 12345
	PRINT #foo


	while 1
	wend

BANK 50
	BITMAP "01234567"
