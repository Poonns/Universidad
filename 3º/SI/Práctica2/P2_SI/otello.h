#ifndef _OTELLO_H
#define _OTELLO_H

#define DEB true

#include <stdlib.h>

#include <iostream>
#include <list>
#include <string>
#include <float.h>

#define BOARD_SIZE 8

using namespace std;

// Color for background

const string BG_COL="246";
const string FG_COL="0";

const int Brigth=60;

// A square is a pair of two characters: the first is the row and must be '1',...,'8'. The second is the column and must be 'A',...,'H' (in upper case).
typedef pair<char,char> Square;
// A list of squares can be useful, for instance, to store the possible places to put next token (possible moves)
typedef list<Square>    Sqlist;

// The two tokens that will be used (like the colors of both sides) are x and o. e means empty square.
enum Token { x, o, e };

// The two players will be named a Bob and Alice. Alice will play with 'o' and Bob with 'x'. Alice will start the game.
enum Player { Hector, Daniel };

// A move is represented as a pair which says who makes the move and in which square he/she will place his/her token.
typedef pair<Player,Square> Move;

// This is the special non-existent square that is to be introduced as second component of a move to represent the fact
// that the player in the first component cannot execute any legal movement, and therefore he/she has to pass. 
const Square pass('0','P');

class Board
{
 public:
	Board();
	Board(bool);
	void Show(ostream &outst);
	Token Content(Square sq);
	bool IsLegal(Move mv);
	void MakeMove(Move mv);
	Board TryMove(Move mv,bool &legal);
	bool GameOver(Player &winner,unsigned &winnerpoints,unsigned &looserpoints);
	Sqlist ValidMoves(Player pl);
 private:
// The next two definitions are only for internal use. Forget about them (and don't use them).
	typedef pair<unsigned,unsigned> internalsq;
	typedef list<internalsq> lsq;	
	Token content[BOARD_SIZE][BOARD_SIZE];
	lsq WhereToLook(internalsq s,unsigned dir);
	bool IsLegalDoing(Move mv,lsq &to_revert);
	void ExecuteMove(Move mv,lsq revert);
	unsigned GetSquares(Player pl);
	void ColChar(ostream &outst,char c);
	void ColString(ostream &outst,string s);
	void BlackDisk(ostream &outst);
	void WhiteDisk(ostream &outst);
};

string Name(Player pl);
Player Opponent(Player pl);
Token Tk(Player pl);
char Symbol(Token tk);
bool ValidSquare(Square sq);
bool ValidPlayer(Player pl);
#endif
