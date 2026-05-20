# Covers
Pico Launcher supports showing covers for any kind of file by placing `.bmp` files in the right folders.

## Where to place covers
Covers are placed in subfolders of the `/_pico/covers` folder.
- The `/_pico/covers/nds` and `/_pico/covers/gba` folders contain DS and GBA covers, respectively, by game code. For example `ABCD.bmp`.
- The `/_pico/covers/user` folder contains covers for specific file names. For example `myGame.nds.bmp` will be used for any file named `myGame.nds`. This takes precedence over a game code based cover. Since most consoles other than DS and GBA don't have game codes, this folder is suited to provide covers for games of other systems, for example NES or GB(C).

## Cover format
The covers should be 8 bpp (256 colors) `.bmp` files of 128x96 pixels.
Only the top-left 106x96 pixels are used. The remaining part on the right is padding and is ignored.

### Example
Note how the black part on the right is not used.

![Cover example](images/CoverExample.bmp)