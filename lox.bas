	REM Lord Of Xmas

BANK ROM 1024

OPTION EXPLICIT

DEF FN PRINT_XY(X, Y, S) = PRINT AT X+Y*32,S

' Tile/Char memory layout:
' Tile 0 -> 32: Common tiles.
' CVBasic loads a default font in the tiles from 32 to 127
' I will likely use a few less useful characters as icons or for translations.
' Tile 128 -> 191: Map tiles.
' Tile 192 -> 255: TBD.
CONST TILE_FONT_MIN = 32
CONST TILE_FONT_MAX = 127
CONST TILE_MAP_MIN  = 128
CONST TILE_MAP_LEN  = 64
CONST TILE_MAP_MAX  = 128 + TILE_MAP_LEN

' Sprite Layout
' First 16 sprites: player.
' Next 16 sprites: player weapon.
' Next 16 sprites: Menu icons. (?)
' Next 16 sprites: Foe projectiles.
' Then each sprite per 16.
CONST SPRITE_PLAYER = 0
CONST SPRITE_WEAPON = 16

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


start:
' Set up screen
	MODE 4
	BORDER 12
	SPRITE FLICKER OFF
	CLS

' Display opening image
	BANK SELECT BACK_BANK_1
	DEFINE CHAR 128,64,nun_char
	SCREEN nun_pattern,0,300,8,8,8
' Show opening message
	BANK SELECT 0
	PRINT_XY(10, 4,"Lord Of Xmas")
	PRINT_XY(6, 5,"By xmasengine, 2025")
' Display player sprite
	player_sprite_x = 150
	player_sprite_y = 40
	player_sprite_f = 0
	player_sprite_d = FRAME_S
	BANK SELECT SPRITE_BANK_1
	GOSUB sprite_palette
	DEFINE SPRITE SPRITE_PLAYER,64,sprite_bitmap
	SPRITE SPRITE_PLAYER,player_sprite_y,player_sprite_x,player_sprite_f
	BANK SELECT MAIN_BANK
	WHILE 1
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
		ELSEIF CONT1.BUTTON > 0 THEN
			BANK SELECT MAP_BANK_1
			DEFINE CHAR 128,64,map1_bitmap
			gosub map1_palette
			BANK SELECT MAIN_BANK
		ELSEIF CONT1.BUTTON2 > 0 THEN
			BANK SELECT MAP_BANK_1
			SCREEN map1_map
			BANK SELECT MAIN_BANK
		END IF
		IF (FRAME % 15) = 0 THEN player_sprite_f = player_sprite_f + 2
		IF player_sprite_f > 3 THEN player_sprite_f = 0
		SPRITE SPRITE_PLAYER, player_sprite_y,player_sprite_x,player_sprite_f+player_sprite_d
		WAIT
	WEND
	SCREEN DISABLE
	WAIT
' halt
done: GOTO done
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
