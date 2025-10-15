	REM Lord Of Xmas

BANK ROM 1024

OPTION EXPLICIT

DEF FN PRINT_XY(X, Y, S) = PRINT AT X+Y*32,S

' Tile/Char memory layout:
' Tile 0 -> 32: HUD tiles.
' CVBasic loads a default font in the tiles from 32 to 127
' I will likely use a few less useful characters as icons or for translations.
' Tile 128 -> 191: Map tiles.
' Tile 192 -> 255: TBD.
CONST TILE_FONT_MIN = 32
CONST TILE_FONT_MAX = 127
CONST TILE_MAP_MIN  = 128
CONST TILE_MAP_LEN  = 64
CONST TILE_MAP_MAX  = 128 + TILE_MAP_LEN

' 8 frames per player, foe, or character, 2 per direction for animation.
' Sprite layout on normal screen: 128 sprites available.
'  0 -   7: Player 1.
'  8 -  15: Player 1 special action.
' 16 -  19: Player 1 weapon.
' 20 -  23: Player 1 projectile.
' 24 -  31: Player 2.
' 32 -  39: Player 2 special action.
' 40 -  43: Player 2 weapon.
' 44 -  47: Player 2 projectile.
' 48 -  64: HUD Icons (16).
' 64 - 128: Presences loaded by the map. Includes foe weapons and projectiles.
'
' Sprite layout on menu screen is different and only has icons.
' Icon 0 is the cursor.
'

CONST SPRITE_PLAYER = 0
CONST SPRITE_WEAPON = 16
CONST SPRITE_MAP    = 64

CONST MAIN_BANK = 0

CONST SPRITE_BANK_1 = 4
CONST SPRITE_BANK_2 = 5
CONST SPRITE_BANK_3 = 6
CONST SPRITE_BANK_4 = 7

CONST TILE_BANK_1 = 8
CONST TILE_BANK_2 = 9
CONST TILE_BANK_3 = 10
CONST TILE_BANK_4 = 11

CONST BACK_BANK_1 = 12
CONST BACK_BANK_2 = 13
CONST BACK_BANK_3 = 14
CONST BACK_BANK_4 = 15

CONST MAP_BANK_1 = 16
CONST MAP_BANK_2 = 17
CONST MAP_BANK_3 = 18
CONST MAP_BANK_4 = 19

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


DIM player_sprite_x
DIM player_sprite_y
DIM player_sprite_f
DIM player_sprite_d
DIM player_sprite_m
DIM player_facing_tx
DIM player_facing_ty
DIM map_flag_peek
DIM map_flag_peek_1
DIM map_flag_peek_2


CONST BORDER_LEFT_ON=1
CONST BORDER_NO_SCROLL_LEFT=2
CONST BORDER_NO_SCROLL_BOTTOM=2

const FLAG_SOLID = 32
const FLAG_HARM  = 64
const FLAG_BLESS = 128

CONST #VRAM_TILE_MAP=$3800

' Reads the map flag from video memory.
' There are 3 extra bits available which Lox uses for collision detection.
' For a sprite X and Y should be adjusted depending on its size.
DEF FN MAP_FLAG_AT(X, Y, VAR) = DEFINE VRAM READ #VRAM_TILE_MAP + ((X) / 8 + ((Y) / 8) * 32)*2 + 1, 1, VAR



start:
' Set up screen
	MODE 4
	BORDER 0, BORDER_NO_SCROLL_BOTTOM + BORDER_LEFT_ON
	SPRITE FLICKER OFF
	CLS

' Display opening image
	BANK SELECT BACK_BANK_1
	DEFINE CHAR 128,64,nun_char
	SCREEN nun_pattern,0,300,8,8,8
' Show opening message
	BANK SELECT 0
	PRINT_XY(10, 3,"Lord Of Xmas")
	PRINT_XY(6, 4,"By xmasengine, 2025")
	PRINT_XY(6, 5,"Press a button")
	PRINT_XY(0, 6," ")
	PLAY SIMPLE
	PLAY music_kyrie_eleison

' Wait for button press
	WHILE CONT.BUTTON = 0 AND CONT.BUTTON2 = 0
	WAIT
	WEND

' Load tile map: select bank, load CHAR bitmaps, load map as screen
	BANK SELECT MAP_BANK_3
	DEFINE CHAR 128,32,m0003church_bitmap
' Should not need parameters as a single map is a single screen for now
	SCREEN m0003church_map
' Palette is loaded though a gosub
	GOSUB m0003church_palette
' Debug
	PRINT_XY(1, 23, "Map Loaded")
' Display player sprite
	player_sprite_x = 120
	player_sprite_y = 120
	player_sprite_f = 0
	player_sprite_d = FRAME_S
	BANK SELECT SPRITE_BANK_1
	GOSUB sprite_palette
	DEFINE SPRITE SPRITE_PLAYER,64,sprite_bitmap
	SPRITE SPRITE_PLAYER,player_sprite_y,player_sprite_x,player_sprite_f
	WAIT

	BANK SELECT MAIN_BANK
	WHILE 1
		player_sprite_m = 0
		map_flag_peek = 0

		IF CONT1.UP THEN
			player_sprite_d = FRAME_N
			MAP_FLAG_AT(player_sprite_x+4, player_sprite_y+7, map_flag_peek)
			MAP_FLAG_AT(player_sprite_x+1, player_sprite_y+7, map_flag_peek_1)
			MAP_FLAG_AT(player_sprite_x+7, player_sprite_y+7, map_flag_peek_2)
			IF map_flag_peek AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_1 AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_2 AND FLAG_SOLID THEN
			ELSE
				player_sprite_y = player_sprite_y - 1
				player_sprite_m = 1
			END IF
		ELSEIF CONT1.DOWN THEN
			player_sprite_d = FRAME_S
			MAP_FLAG_AT(player_sprite_x+4, player_sprite_y+17, map_flag_peek)
			MAP_FLAG_AT(player_sprite_x+1, player_sprite_y+17, map_flag_peek_1)
			MAP_FLAG_AT(player_sprite_x+7, player_sprite_y+17, map_flag_peek_2)
			IF map_flag_peek AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_1 AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_2 AND FLAG_SOLID THEN
			ELSE
				player_sprite_y = player_sprite_y + 1
				player_sprite_m = 1
			END IF
		END IF

		IF CONT1.LEFT THEN
			player_sprite_d = FRAME_W
			MAP_FLAG_AT(player_sprite_x-1, player_sprite_y+12, map_flag_peek)
			MAP_FLAG_AT(player_sprite_x-1, player_sprite_y+9, map_flag_peek_1)
			MAP_FLAG_AT(player_sprite_x-1, player_sprite_y+14, map_flag_peek_2)
			IF map_flag_peek AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_1 AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_2 AND FLAG_SOLID THEN
			ELSE
				player_sprite_x = player_sprite_x - 1
				player_sprite_m = 1
			END IF
		ELSEIF CONT1.RIGHT THEN
			player_sprite_d = FRAME_E
			MAP_FLAG_AT(player_sprite_x+9, player_sprite_y+12, map_flag_peek)
			MAP_FLAG_AT(player_sprite_x+9, player_sprite_y+9, map_flag_peek_1)
			MAP_FLAG_AT(player_sprite_x+9, player_sprite_y+14, map_flag_peek_2)
			IF map_flag_peek AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_1 AND FLAG_SOLID THEN
			ELSEIF map_flag_peek_2 AND FLAG_SOLID THEN
			ELSE
				player_sprite_x = player_sprite_x + 1
				player_sprite_m = 1
			END IF
		END IF

		IF CONT1.BUTTON > 0 THEN
			player_sprite_m = 1
		ELSEIF CONT1.BUTTON2 > 0 THEN
			player_sprite_m = 1
		END IF
		IF player_sprite_m > 0 THEN
			IF (FRAME % 15) = 0 THEN player_sprite_f = player_sprite_f + 2
			IF player_sprite_f > 3 THEN player_sprite_f = 0
		END IF
		SPRITE SPRITE_PLAYER, player_sprite_y,player_sprite_x,player_sprite_f+player_sprite_d
' DEFINE VRAM READ #VRAM_TILE_MAP + ((player_sprite_x) / 8 + ((player_sprite_y+8) / 8) * 32)*2 + 1, 1, map_flag_peek
		PRINT_XY(12, 23, map_flag_peek)
		PRINT_XY(16, 23, "   ")
		IF map_flag_peek AND FLAG_SOLID THEN PRINT_XY(16, 23, "S")
		IF map_flag_peek AND FLAG_HARM THEN PRINT_XY(17, 23, "H")
		IF map_flag_peek AND FLAG_BLESS THEN PRINT_XY(18, 23, "B")
		WAIT
	WEND
	SCREEN DISABLE
	WAIT
' halt
done: GOTO done
	music_0: DATA BYTE 32
	MUSIC D4F
	MUSIC S
	MUSIC C4
	MUSIC D4
	MUSIC S
	MUSIC -
	MUSIC D4F,D4Z
	MUSIC S,C4
	MUSIC C4,C4
	MUSIC D4,D4
	MUSIC S,S
	MUSIC -,S
	MUSIC D4
	MUSIC S
	MUSIC E4
	MUSIC F4
	MUSIC S
	MUSIC -
	MUSIC D4,D4
	MUSIC S,E4
	MUSIC E4,G4
	MUSIC F4,F4
	MUSIC S,S
	MUSIC -,S
	MUSIC F4
	MUSIC S
	MUSIC G4
	MUSIC S
	MUSIC D4
	MUSIC S
	MUSIC F4,-
	MUSIC S,F4
	MUSIC G4,S
	MUSIC S,G4
	MUSIC D4,S
	MUSIC S,D4
	MUSIC C4
	MUSIC D4
	MUSIC A4
	MUSIC A4
	MUSIC G4
	MUSIC -
	MUSIC F4
	MUSIC S
	MUSIC G4
	MUSIC S
	MUSIC D4
	MUSIC S
	MUSIC F4
	MUSIC E4
	MUSIC F4
	MUSIC E4
	MUSIC D4
	MUSIC -
	MUSIC A4
	MUSIC S
	MUSIC G4
	MUSIC S
	MUSIC F4
	MUSIC S
	MUSIC E4
	MUSIC S
	MUSIC F4
	MUSIC S
	MUSIC -
	MUSIC -
	MUSIC REPEAT
    MUSIC STOP
	include "./music/kirie.bas"

' Sprite bank
	BANK 4
	INCLUDE "sprite1.bas"
' Tile bank 2
	BANK TILE_BANK_1
	INCLUDE "tile2.bas"
' Intro image bank
	BANK BACK_BANK_1
	INCLUDE "nun.bas"
	BANK MAP_BANK_1
	INCLUDE "map1.bas"
    BANK MAP_BANK_3
    INCLUDE "./map/m0003-church.xml.bas"

