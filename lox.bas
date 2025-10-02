	REM Lord Of Xmas

BANK ROM 1024

OPTION EXPLICIT

DEF FN PRINT_XY(X, Y, S) = PRINT AT X+Y*32,S

CONST FPS = 60

CONST FRAME_S = 0
CONST FRAME_S_1 = 0
CONST FRAME_S_2 = 2

CONST FRAME_E = 4
CONST FRAME_E_1 = 4
CONST FRAME_E_2 = 6

CONST FRAME_N = 8
CONST FRAME_N_1 = 8
CONST FRAME_N_2 = 10

CONST FRAME_W = 12
CONST FRAME_W_1 = 12
CONST FRAME_W_2 = 14


CONST player_sprite_i = 0
DIM player_sprite_x
DIM player_sprite_y
DIM player_sprite_f
DIM player_sprite_d


start:
' Set up screen
	MODE 4
	BORDER 12
	CLS

' Display opening image
	BANK SELECT 10
	DEFINE CHAR 128,64,nun_char
	SCREEN nun_pattern,0,300,8,8,8
' Display player sprite
	player_sprite_x = 150
	player_sprite_y = 40
	player_sprite_f = 0
	player_sprite_d = FRAME_S
	BANK SELECT 3
	DEFINE SPRITE 0,64,sprite_bitmap
	GOSUB sprite_palette
	SPRITE player_sprite_i,player_sprite_y,player_sprite_x,player_sprite_f
' Show opening message
	BANK SELECT 0
	PRINT_XY(10, 4,"Lord Of Xmas")
	PRINT_XY(6, 5,"By xmasengine, 2025")
	BANK SELECT 0
	WHILE CONT1.BUTTON = 0
		IF CONT1.UP THEN
			player_sprite_y = player_sprite_y - 1
			player_sprite_d = FRAME_N
		ELSEIF CONT1.DOWN THEN
			player_sprite_y = player_sprite_y + 1
			player_sprite_d = FRAME_S
		ELSEIF CONT1.LEFT THEN
			player_sprite_x = player_sprite_x - 1
			player_sprite_d = FRAME_W
		ELSEIF CONT1.RIGHT THEN
			player_sprite_x = player_sprite_x + 1
			player_sprite_d = FRAME_E
		END IF
		IF (FRAME % 15) = 0 THEN player_sprite_f = player_sprite_f + 2
		IF player_sprite_f > 3 THEN player_sprite_f = 0
		SPRITE player_sprite_i, player_sprite_y,player_sprite_x,player_sprite_f+player_sprite_d
		WAIT
	WEND
	SCREEN DISABLE
	WAIT
' halt
done: GOTO done
' Sprite bank
	BANK 3
	INCLUDE "sprite1.bas"
' Intro image bank
	BANK 10
	INCLUDE "nun.bas"
