	REM Lord Of Xmas

OPTION EXPLICIT

DEF FN PRINT_XY(X, Y, S) = PRINT AT X+Y*32,S

start:
	MODE 4
	CLS


	' Display image.
	' Display image.
	MODE 4
	DEFINE CHAR 128,64,nun_char
	SCREEN nun_pattern,0,0,8,8,8
	' Show message.
	PRINT_XY(10, 4,"Lord Of Xmas")
	PRINT_XY(6, 5,"By xmasengine, 2025")
done: GOTO done
	INCLUDE "nun.bas"

