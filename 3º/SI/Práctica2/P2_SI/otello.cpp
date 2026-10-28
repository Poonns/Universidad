#include "otello.h"

string Name(Player pl)
{
 switch (pl)
 {
  case Daniel: return "Daniel"; break;
  case Hector: return "Hector"; break;
  default: return "Unknown player"; break;
 }
}

Player Opponent(Player pl)
{
 switch (pl)
 {
  case Daniel: return Hector; break;
  case Hector: return Daniel; break;
  default: cerr << "Error from Opponent: unknown player.\n"; exit(1); break;
 }
}

Token Tk(Player pl)
{
 switch (pl)
 {
  case Daniel: return o; break;
  case Hector: return x; break;
  default: cerr << "Warning: call to function Tk with invalid player.\n"; return e; break;
 }
}

char Symbol(Token tk)
{
 switch (tk)
 {
  case x: return 'x'; break;
  case o: return 'o'; break;
  case e: return ' '; break;
  default: cerr << "Warning: call to function Symbol with invalid token.\n"; return ' '; break;
 }
}

bool ValidSquare(Square sq)
{
 return ((sq.first>='1') && (sq.first<='8') && (sq.second>='A') && (sq.second<='H'));
}

bool ValidPlayer(Player pl)
{
 return ((pl==Daniel) || (pl==Hector));
}

Sqlist Board::ValidMoves(Player pl)
{
 Sqlist ret;

 for (unsigned i=0;i<BOARD_SIZE;i++)
  for (unsigned j=0;j<BOARD_SIZE;j++)
  {
   Square here('1'+i,'A'+j);
   Move mv(pl,here);
   if ( (Content(here)==e) && (IsLegal(mv)) )
    ret.push_back(here);
  }

 return ret;
}

unsigned Board::GetSquares(Player pl)
{
 Token t=Tk(pl);

 unsigned points=0;
 for (unsigned i=0;i<BOARD_SIZE;i++)
  for (unsigned j=0;j<BOARD_SIZE;j++)
   if (content[i][j]==t)
    points++;

 return points;
}

bool Board::GameOver(Player &winner,unsigned &winnerpoints,unsigned &looserpoints)
{
 Sqlist Alicemoves=ValidMoves(Hector);
 Sqlist Bobmoves=ValidMoves(Daniel);
 
 if (Alicemoves.empty() && Bobmoves.empty())
 {
  unsigned Alicepoints=GetSquares(Hector);
  unsigned Bobpoints=GetSquares(Daniel);
  if (Alicepoints>=Bobpoints)
  {
   winnerpoints=Alicepoints;
   looserpoints=Bobpoints;
   winner=Hector;
  }
  else
  {
   winnerpoints=Bobpoints;
   looserpoints=Alicepoints;
   winner=Daniel;
  }
  return true;
 }
 else
  return false;
}

Board::Board()
{
 for (unsigned i=0;i<BOARD_SIZE;i++)
  for (unsigned j=0;j<BOARD_SIZE;j++)
   content[i][j]=e;
 content[3][3]=content[4][4]=o;
 content[3][4]=content[4][3]=x;
}

Board::Board(bool error)
{
 for (unsigned i=0;i<BOARD_SIZE;i++)
  for (unsigned j=0;j<BOARD_SIZE;j++)
   content[i][j]=e;
}

void Board::Show(ostream &outst)
{
 outst << "\n";
 ColString(outst,"    ");
  
 for (unsigned i=0;i<BOARD_SIZE;i++)
 {
  ColChar(outst,' ');
  ColChar(outst,'A'+i);
  ColString(outst," +");
 }
 outst << endl;
 ColString(outst,"   +");

 for (unsigned i=0;i<BOARD_SIZE;i++)
 {
  for (unsigned j=0;j<BOARD_SIZE;j++)
   ColString(outst,"---+");
  outst << endl;
  ColChar(outst,' ');
  ColChar(outst,char('1'+i));
  ColString(outst," |");
  for (unsigned j=0;j<BOARD_SIZE;j++)
  {
   ColChar(outst,' ');
   switch (content[i][j])
   {
    case x: BlackDisk(outst); break;
    case o: WhiteDisk(outst); break;
    case e: ColChar(outst,' ');; break;
   }
   ColString(outst," |");
  }
  outst << endl;
  ColString(outst,"   +");
 }
 for (unsigned j=0;j<BOARD_SIZE;j++)
  ColString(outst,"---+");
 outst << endl << endl;
}

Token Board::Content(Square sq)
{
 if (!ValidSquare(sq))
 {
  cerr << "Error from Board::Content: invalid square passed.\n";
  exit(1);
 }

 unsigned row=unsigned(sq.first-'1');
 unsigned col=unsigned(sq.second-'A');

 return content[row][col];
}

Board::lsq Board::WhereToLook(internalsq s,unsigned dir)
{
 lsq ret;
 unsigned r=s.first;
 unsigned c=s.second;
 internalsq current;

 switch (dir)
 {
  case 0: while (                    (c>0))            {      c--; current.first=r; current.second=c; ret.push_back(current); }; break;
  case 1: while ((r>0)            && (c>0))            { r--; c--; current.first=r; current.second=c; ret.push_back(current); }; break;
  case 2: while ((r>0))                                { r--;    ; current.first=r; current.second=c; ret.push_back(current); }; break;
  case 3: while ((r>0)            && (c<BOARD_SIZE-1)) { r--; c++; current.first=r; current.second=c; ret.push_back(current); }; break;
  case 4: while (                    (c<BOARD_SIZE-1)) {      c++; current.first=r; current.second=c; ret.push_back(current); }; break;
  case 5: while ((r<BOARD_SIZE-1) && (c<BOARD_SIZE-1)) { r++; c++; current.first=r; current.second=c; ret.push_back(current); }; break;
  case 6: while ((r<BOARD_SIZE-1))                     { r++;      current.first=r; current.second=c; ret.push_back(current); }; break;
  case 7: while ((r<BOARD_SIZE-1) && (c>0))            { r++; c--; current.first=r; current.second=c; ret.push_back(current); }; break;
  default: { cerr << "Error from Board::WhereToLook: direction is a number between 0 and 7.\n"; exit(1); }; break;
 }

 return ret;
}

bool Board::IsLegalDoing(Move mv,lsq &to_revert)
{
 if (!to_revert.empty())
 {
  cerr << "Error from Board::IsLegalDoing: list of squares to revert is supposed to be passed empty.\n";
  exit(1);
 }
 if (!ValidSquare(mv.second))
 {
  cerr << "Error from Board::IsLegalDoing: second (square) component of passed move is invalid.\n";
  exit(1);
 }
 if (!ValidPlayer(mv.first))
 {
  cerr << "Error from Board::IsLegalDoing: first (player) component of passed move is invalid.\n";
  exit(1);
 }

 unsigned row=unsigned(mv.second.first-'1');
 unsigned col=unsigned(mv.second.second-'A');

 if (content[row][col]!=e)
  return false;
 
 internalsq s(row,col);
 Token mycol=Tk(mv.first);
 Token oppcol=Tk(Opponent(mv.first));
 for (unsigned d=0;d<8;d++)
 {
  lsq l=WhereToLook(s,d);
  lsq::iterator it=l.begin();
  while (it!=l.end() && (content[it->first][it->second]==oppcol))
   it++;
  if (it!=l.end() && content[it->first][it->second]==mycol)
  {
   for (lsq::iterator it2=l.begin();it2!=it;it2++)
   {
    internalsq revsq(it2->first,it2->second);
    to_revert.push_back(revsq);
   }
  }
 }

 return (to_revert.empty() ? false : true);
}

bool Board::IsLegal(Move mv)
{
 lsq r;
 return IsLegalDoing(mv,r);
}

void Board::ExecuteMove(Move mv,lsq revert)
{
 Token to_put=Tk(mv.first);

 unsigned row=unsigned(mv.second.first-'1');
 unsigned col=unsigned(mv.second.second-'A');

 content[row][col]=to_put;
 for (lsq::iterator it=revert.begin();it!=revert.end();it++)
  content[it->first][it->second]=to_put;
}

void Board::MakeMove(Move mv)
{
 // First, if the movement is pass, do nothing.
 if (mv.second==pass)
  return;

 lsq revert;
 if (IsLegalDoing(mv,revert))
  ExecuteMove(mv,revert);
}
	
Board Board::TryMove(Move mv,bool &legal)
{
 lsq revert;
 if (IsLegalDoing(mv,revert))
 {
  legal=true;
  Board b=(*this);
  b.ExecuteMove(mv,revert);
  return b;
 }
 else
 {
  legal=false;
  Board b(true);
  return b;
 }
}

void Board::WhiteDisk(ostream &outst)
{
 outst << "\033[38;5;0;48;5;15m@\033[0m";
}

void Board::BlackDisk(ostream &outst)
{
 outst << "\033[38;5;15;48;5;0m@\033[0m";
}

void Board::ColChar(ostream &outst,char c)
{
 outst << "\033[38;5;"<<FG_COL<<";48;5;"<<BG_COL<<"m" << c << "\033[0m";
}


void Board::ColString(ostream &outst,string s)
{
 for (unsigned i=0;i<s.length();i++)
  outst << "\033[38;5;"<<FG_COL<<";48;5;"<<BG_COL<<"m" << s[i] << "\033[0m";
}
