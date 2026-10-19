#include "takuzu.h"

Board::Board(unsigned size)
{
 for (unsigned r=0; r<size; r++)
 {
  vector<Token> thisrow;
  for (unsigned c=0; c<size; c++)
   thisrow.push_back(e);
  content.push_back(thisrow);
 }
}

Board::Board(string file_name,ostream &outst)
{
 *this=Board(file_name);
 cout << "Initial board read from file " << file_name << ":\n";
 Show(outst);
 cout << "Initial board after setting the compulsory moves:\n";
 SetAllCompulsory();
 Show(outst);
}

Board::Board(string file_name)
{
 ifstream f(file_name.c_str());
 if (!f.is_open())
 {
  cerr << "Error in Board contructor: file " << file_name << " does not exist or cannot be opened.\n";
  exit(1);
 }
 
 string row;
 
 f >> row;
 size = row.size();
 if ((size % 2) != 0)
 {
  cerr << "Error in Board contructor when reading from file " << file_name << ": board size must be even.\n";
  exit(1);
 }
 
 unsigned r,c;
 for (r=0; r<size; r++)
 {
  vector<Token> thisrow;
  content.push_back(thisrow);
 }
 
 r=0;
 while (r < size)
 {
  if (row.size()!=size)
  {
   cerr << "Error in Board constructor reading file " << file_name << ":\n";
   cerr << "  row " << row << " has " << row.size() << " characters instead of " << size << ".\n";
   exit(1);
  }
  c=0;
  while (c<size)
  {
   switch (row[c])
   {
    case 'b':
    case 'B': content[r].push_back(b); break;
    case 'w':
    case 'W': content[r].push_back(w); break;
    case 'e':
    case 'E': content[r].push_back(e); break;
    default: 
    {
     cerr << "Error in Board constructor reading file " << file_name << ":\n";
     cerr << "  Character at row " << r << ", column " << c+1 << " is not 'w','W','b','B','e' or 'E'.\n";
     cerr << "  No other characters are allowed.\n";
     exit(1);
    }
   }
   c++;
  } 
  r++;
  if (r<size)
   f >> row;
 }
 f.close();
}

Token Board::Content(Position p)
{
 if ((p.row>=size) || (p.col>=size))
 {
  cerr << "Error in Board::Content: requested row or column are outside the board.\n";
  exit(1);
 }
 return content[p.row][p.col];
}

void Board::WhiteDisk(ostream &outst)
{
 outst << "\033[38;5;0;48;5;15mw\033[0m";
}

void Board::BlackDisk(ostream &outst)
{
 outst << "\033[38;5;15;48;5;0mb\033[0m";
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

void Board::Show(ostream &outst)
{
 outst << "\n";
 
 ColString(outst,"     "); 
 for (unsigned i=0;i<size;i++)
 {
  ColChar(outst,' ');
  ColChar(outst,'0'+(i%10));
  ColString(outst," |");
 }
 outst << endl;
 
 ColString(outst,"    |"); 
 for (unsigned i=0;i<size;i++)
 {
  ColChar(outst,' ');
  if (i>9)
   ColChar(outst,'0'+i);
  else
   ColChar(outst,' ');
  ColString(outst," |");
 }
 outst << endl;
 
 ColString(outst,"    +");

 for (unsigned i=0;i<size;i++)
 {
  for (unsigned j=0;j<size;j++)
   ColString(outst,"---+");
  outst << endl;
  ColChar(outst,' ');
  if (i<10)
   ColChar(outst,' ');
  else
   ColChar(outst,char('0'+(i/10)));
  ColChar(outst,char('0'+i));
  ColString(outst," |");
  for (unsigned j=0;j<size;j++)
  {
   ColChar(outst,' ');
   switch (content[i][j])
   {
    case e: ColChar(outst,' ');; break;
    case b: BlackDisk(outst); break;
    case w: WhiteDisk(outst); break;
   }
   ColString(outst," |");
  }
  outst << endl;
  ColString(outst,"    +");
 }
 for (unsigned j=0;j<size;j++)
  ColString(outst,"---+");
 outst << endl << endl;
}

bool Board::PossibleLeft(unsigned r,unsigned c,Token t)
{
 if (c<2)
  return true;
 if (content[r][c-2]==t && content[r][c-1]==t)
  return false;
 return true;
}

bool Board::PossibleRight(unsigned r,unsigned c,Token t)
{
 if (c>=size-2)
  return true;
 if (content[r][c+1]==t && content[r][c+2]==t)
  return false;
 return true;
}

bool Board::PossibleUp(unsigned r,unsigned c,Token t)
{
 if (r<2)
  return true;
 if (content[r-2][c]==t && content[r-1][c]==t)
  return false;
 return true;
}

bool Board::PossibleDown(unsigned r,unsigned c,Token t)
{
 if (r>=size-2)
  return true;
 if (content[r+1][c]==t && content[r+2][c]==t)
  return false;
 return true;
}

bool Board::PossibleCenterHor(unsigned r,unsigned c,Token t)
{
 if ((r==0) || (r==size-1))
  return true;
 if (content[r-1][c]==t && content[r+1][c]==t)
  return false;
 return true;
}

bool Board::PossibleCenterVer(unsigned r,unsigned c,Token t)
{
 if ((c==0) || (c==size-1))
  return true;
 if (content[r][c-1]==t && content[r][c+1]==t)
  return false;
 return true;
}

bool Board::PossibleCount(unsigned r,unsigned c,Token t)
{
 unsigned numtype=0;
 for (unsigned r1=0; r1<size; r1++)
  if (content[r1][c]==t)
   numtype++;
 if (numtype+1>(size/2))
  return false;
 numtype=0;
 for (unsigned c1=0; c1<size; c1++)
  if (content[r][c1]==t)
   numtype++;
 if (numtype+1>(size/2))
  return false;
 return true;
}

bool Board::PossibleRep(unsigned r,unsigned c,Token t)
{
 Token oldtoken=content[r][c];
 content[r][c]=t;
  
 bool eqrow;
 unsigned ir,ir2;
 unsigned qc;
 ir=0;
 while (ir<size-1)
 {
  ir2=ir+1;
  while (ir2<size)
  {
   eqrow=true;
   qc=0;
   while (qc<size && eqrow)
   {
    eqrow = (content[ir][qc]==content[ir2][qc]) && (content[ir][qc]!=e);
    qc++;
   }
   if (eqrow)
   {
    content[r][c]=oldtoken;
    return(false);
   }
   ir2++;
  }
  ir++;
 }
 
 bool eqcol;
 unsigned ic,ic2;
 unsigned qr;
 ic=0;
 while (ic<size-1)
 {
  ic2=ic+1;
  while (ic2<size)
  {
   eqcol=true;
   qr=0;
   while (qr<size && eqcol)
   {
    eqcol = (content[qr][ic]==content[qr][ic2]) && (content[qr][ic]!=e);
    qr++;
   }
   if (eqcol)
   {
    content[r][c]=oldtoken;
    return(false);
   }
   ic2++;
  }
  ic++;
 }
 
 content[r][c]=oldtoken;
 return(true);
}

bool Board::Possible(Move m)
{
 if ((m.sq.row>=size) || (m.sq.col>=size))
 {
  cerr << "Error in Board::Possible: requested row or column are outside the board.\n";
  exit(1);
 }
 if (m.value==e)
 {
  cerr << "Error in Board::Possible: third argument must be w or b, never e\n";
  exit(1);
 }
 
 // If the requested place is not empty, it is obviously impossible to put anything on it
 if (content[m.sq.row][m.sq.col]!=e)
  return(false);
 
 bool ret = PossibleLeft(m.sq.row,m.sq.col,m.value) &&
            PossibleRight(m.sq.row,m.sq.col,m.value) &&
            PossibleUp(m.sq.row,m.sq.col,m.value) &&
            PossibleDown(m.sq.row,m.sq.col,m.value) &&
            PossibleCenterHor(m.sq.row,m.sq.col,m.value) &&
            PossibleCenterVer(m.sq.row,m.sq.col,m.value) &&
            PossibleCount(m.sq.row,m.sq.col,m.value) &&
            PossibleRep(m.sq.row,m.sq.col,m.value);

 return(ret);
}

void Board::MakeMove(Move m)
{
 content[m.sq.row][m.sq.col]=m.value;
}

bool Board::Compulsory(Move &m)
{
 if ((m.sq.row>=size) || (m.sq.col>=size))
 {
  cerr << "Error in Board::MustBe: requested row or column are outside the board.\n";
  exit(1);
 }
 
 // If the place is not empty, obviously nothing can be put there so it cannot be compulsory
 if (content[m.sq.row][m.sq.col]!=e)
  return(false);
  
 // Let's build two movements putting different things at the same place, a 'b' and a 'w' piece....
 Move mw=m;
 mw.value=w;
 
 Move mb=m;
 mb.value=b;
 
 // ... and let's see which movement/s is/are possible
 bool pw=Possible(mw);
 bool pb=Possible(mb);
 
 // If both are possible, or none is possible, we haven'nt a compulsory movement:
 // we would have either an optional movement, or no movement at all. We return false.
 if ((pw && pb) || ((!pw) && (!pb)))
  return(false);

 // Otherwise, if we are here, exactly one of the choices is possible and the other is not possible.
 // The possible choice then becomes compulsory. We take note of it (in m) and return true
 m.value =  (pw ? w : b);
 return(true);
}

void Board::SetAllCompulsory()
{
 bool something_changed;
 Move m;

 do
 {
  something_changed=false;
  for (unsigned r=0; r<size; r++)
  {
   m.sq.row=r;
   for (unsigned c=0; c<size; c++)
   {
    m.sq.col=c;
    if ( (content[r][c]==e) && Compulsory(m) )
    {
     something_changed=true;
     MakeMove(m);
    }
   }
  }
 } 
 while (something_changed);
}

void Board::GetNeighbours(Board::lnei &listnei)
{
 Move m;

 // Double loop to run through all squares:
 for (unsigned r=0; r<size; r++)
 {
  m.sq.row=r;
  for (unsigned c=0; c<size; c++)
  {
   m.sq.col=c;
   // If this square is empty we can try a move on it
   if (content[r][c]==e)
   {
    // Let's try first setting a black 
    m.value=b;
    if (Possible(m))
    {
     // If possible, make a copy of the current board
     Board nei_board=*this;
     // Make really the move on the copy
     nei_board.MakeMove(m);
     // Add all other pieces that have to be set due to this move...
     nei_board.SetAllCompulsory();
     // and add the resulting board to the list of neighbours
     listnei.push_back(pair(m,nei_board));
    } // If black was possible
    // Now, let's try setting a white
    m.value=w;
    if (Possible(m))
    {
     Board nei_board=*this;
     nei_board.MakeMove(m);
     nei_board.SetAllCompulsory();
     listnei.push_back(pair(m,nei_board));
    } // if white was possible
   }  // if empty
  }
 }
}

string Board::GetIdent()
{
 string id(size*size+1,'?');
 
 id[size*size]='\0';
 
 for (unsigned r=0; r<size; r++)
  for (unsigned c=0; c<size; c++)
   switch (content[r][c])
   {
    case b:    id[r*size+c]='b'; break;
    case w:    id[r*size+c]='w'; break;
    case e:    id[r*size+c]='e'; break;
    default:   id[r*size+c]='?'; break;
   }
 return id;   
}

bool Board::Member(Board::lb &set)
{
 Board::lb::iterator it=set.begin();
 bool found=false;
 while (!found && it!=set.end())
 {
  found = (it->GetIdent()==GetIdent());
  it++;
 }
 return found;
}

float Scores::GetfValue(Board b)
{
 string ident=b.GetIdent();
 
 return( fsc.contains(ident) ? fsc[ident] : FLT_MAX ); 
}

void Scores::SetfValue(Board b,float v)
{
 fsc[b.GetIdent()]=v;
}

float Scores::GetgValue(Board b)
{
 string ident=b.GetIdent();
 
 return( gsc.contains(ident) ? gsc[ident] : FLT_MAX ); 
}

void Scores::SetgValue(Board b,float v)
{
 gsc[b.GetIdent()]=v;
}


