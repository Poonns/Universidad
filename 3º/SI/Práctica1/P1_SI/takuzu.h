#ifndef _TAKUZU_H
#define _TAKUZU_H

#include <stdlib.h>

#include <iostream>
#include <fstream>
#include <string>

#include <vector>
#include <utility>
#include <list>
#include <float.h>
#include <map>

using namespace std;

#define DEB false

// Maximum board side
const unsigned int MAX_BSIDE=20;

// Colors for background and foreground
const string BG_COL="246";
const string FG_COL="0";

const int Brigth=60;

// The three tokens that will be used are b, w, e to indicate black, white and empty 
enum Token { b, w, e };

// A position is just a square of the board
typedef struct Position
{
 unsigned row;
 unsigned col;
} Position;

// A move is a square, and the token you want to set on it
typedef struct Move
{
 Position sq;
 Token    value;
} Move;

/*
 Class to represent the takuzu board
*/
class Board
{
 public:
        typedef list<pair<Move,Board>> lnei;    // A list of pairs movement-board (states of the game) will be a very used structure, indeed in the public functions of this class
        typedef list<Board> lb;                 // A list of boards, useful to store the open, closed and solution sets
        Board(unsigned size);                   // Constructor with just a size. The board will be left totally empty
	Board(string file_name);                // Constructor to read the initial board state from an ASCII text file
	Board(string file_name,ostream &outst); // As before, but also shows in the terminal (cout/cerr) or in an open file the initial and initial-after-compulsory state of the board. 
// No dynamic structures, so copy constructor and assigment operator left by default
	unsigned GetSize() { return size; };    // Gets the size (as number of squares in the side) of the board
	Token Content(Position p);              // Returns the token (w,b or e) at position p
	void Show(ostream &outst);              // Shows the current state of the board in the terminal using ANSII codes. Argument can be cout, cerr or an open stream
	void GetNeighbours(lnei &listnei);      // Sets the passed variable (by reference) with the list of possible neighbours, i.e. boards that can be reached from the current state
                                                // WARNING: a niehgbour is not only the board with the addition of a new pin but the board with the addition of a new pin, AND
                                                // all pins added forced by the new board state.  
	bool Member(lb &listboards);            // Says if this board is or not a member of the passed list of boards (i.e.: Am I into this list?)
	friend class Scores;

 protected:
        string GetIdent();                      // Gets a unique string wich acts as identifier of the current state of the board		
 private:
	vector<vector<Token>> content;
	unsigned size;
	bool PossibleLeft(unsigned r,unsigned c,Token t);
	bool PossibleRight(unsigned r,unsigned c,Token t);
	bool PossibleUp(unsigned r,unsigned c,Token t);
	bool PossibleDown(unsigned r,unsigned c,Token t);
	bool PossibleCenterHor(unsigned r,unsigned c,Token t);
	bool PossibleCenterVer(unsigned r,unsigned c,Token t);
	bool PossibleCount(unsigned r,unsigned c,Token t);
	bool PossibleRep(unsigned r,unsigned c,Token t);
	void ColChar(ostream &outst,char c);
        void ColString(ostream &outst,string s);
        void BlackDisk(ostream &outst);
        void WhiteDisk(ostream &outst);
        void BlackFixedDisk(ostream &outst);
        void WhiteFixedDisk(ostream &outst);
        bool Possible(Move m);                 // Returns true if the proposed movement can be executed, according to the rules of the game, in the current state of the board
        bool Compulsory(Move &m);              // Move is passed by reference with a fixed position. The token field of m is filled with w or b if the rules of the
	                                       //  game forces this square to be filled as such. Then, true is returned. If there is no compulsory token to place at that
	                                       //  position, false is returned (and the token part of m is left unspecified
	void MakeMove(Move m);                 // Effectively makes the proposed move, changing the internal state
	void SetAllCompulsory();               // Effectively makes all the movements which are obligued by the rules in the current board state
};

class Scores
{
 public:
  float GetfValue(Board b);
  void SetfValue(Board b,float v);
  float GetgValue(Board b);
  void SetgValue(Board b,float v);
 private:
  std::map<string,float> fsc;
  std::map<string,float> gsc;
};

#endif
