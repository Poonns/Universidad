// The Nature of Code
// Daniel Shiffman
// http://natureofcode.com


LSystem lsys;
Turtle turtle;

void setup() {
  size(600, 600);
  
   //TAREA 1
  
   //Isla de Koch
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F-F+F+FF-F-F+F");
   lsys = new LSystem("F-F-F-F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height/3, radians(90));
   */
   
   //Curva del copo de nieve (modificada)
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F+F-F-F+F");
   lsys = new LSystem("-F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height/3, radians(90));
   */
   
   //Isla
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('F', "F+f-FF+F+FF+Ff+FF-f+FF-F-FF-Ff-FFF");
   ruleset[1] = new Rule('f', "ffffff");
   lsys = new LSystem("F+F+F+F", ruleset);
   turtle = new Turtle(lsys.getSentence(), width/10, radians(90));
   */
   
   //Cuadrado
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "FF-F-F-F-FF");
   lsys = new LSystem("F-F-F-F", ruleset);
   turtle = new Turtle(lsys.getSentence(), width/3, radians(90));
   */
   
   //Aspa
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "FF-F-F-F-F-F+F");
   lsys = new LSystem("F-F-F-F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height/3, radians(90));
   */
   
   //Curva de Levy
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "+F--F+");
   lsys = new LSystem("F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height, radians(45));
   */
   
   //Pentaplexity
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F++F++F-----F--F++F");
   lsys = new LSystem("F++F++F++F++F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height/3, radians(36));
   */
   
   //Curva de Hilbert
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('A', "+BF-AFA-FB+");
   ruleset[1] = new Rule('B', "-AF+BFB+FA-");
   lsys = new LSystem("A", ruleset);
   turtle = new Turtle(lsys.getSentence(), height, radians(90));
   */
   
   //ACTIVIDAD 2
   
   //Triángulo de Sierpinski
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('F', "F-G+F+G-F");
   ruleset[1] = new Rule('G', "GG");
   lsys = new LSystem("F-G-G", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 2, radians(120));
   */

   //Curva del Dragón
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('X', "X+YF+");
   ruleset[1] = new Rule('Y', "-FX-Y");
   lsys = new LSystem("FX", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 2, radians(90));
   */

   //TAREA 2
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('L', "LR");
   ruleset[1] = new Rule('R', "--LL++RR++");
   lsys = new LSystem("L", ruleset);
   turtle = new Turtle(lsys.getSentence(), height/3, radians(60));
   */
   
   //ACTIVIDAD 3
   
   //Apartado a)
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('L', "R+L-L+R");
   ruleset[1] = new Rule('R', "L-R+R-L");
   lsys = new LSystem("L", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(60));
   */
   
   //Apartado b)
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('L', "L+L-L-L+L");
   lsys = new LSystem("-L", ruleset);
   turtle = new Turtle(lsys.getSentence(), height/3, radians(90));
   */
   
   //TAREA 3:
   
   //Arbusto a
   
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F[+F]F[-F]F");
   lsys = new LSystem("F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(25.7));
   
   
   //Arbusto b
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F[+F]F[-F][F]");
   lsys = new LSystem("F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(20));
   */

   //Arbusto c
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "FF-[-F+F+F]+[+F-F-F]");
   lsys = new LSystem("F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(22.5));
   */  

   //Arbusto d
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('X', "F[+X]F[-X]+X");
   ruleset[1] = new Rule('F', "FF");
   lsys = new LSystem("X", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(20));
   */

   //Arbusto e
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('X', "F[+X]F[-X]+FX");
   ruleset[1] = new Rule('F', "FF");
   lsys = new LSystem("X", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(25.7));
   */

   //Arbusto f
   /*
   Rule[] ruleset = new Rule[2];
   ruleset[0] = new Rule('X', "F-[[X]+X]+F[+FX]-X");
   ruleset[1] = new Rule('F', "FF");
   lsys = new LSystem("X", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(25.7));
   */
   
   //ACTIVIDAD 5

   //Gramatica 1
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F[+F]F[-F]F[++F]F[--F]");
   lsys = new LSystem("F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(15));
   */
   
   //Gramatica 2
   /*
   Rule[] ruleset = new Rule[1];
   ruleset[0] = new Rule('F', "F[+F]F[-F]F[+F][-F]");
   lsys = new LSystem("F", ruleset);
   turtle = new Turtle(lsys.getSentence(), height / 3, radians(25.7));
   */
}

void draw() {
  background(255);  
  fill(0);
  //text("Click mouse to generate", 10, height-10);

  translate(width/3, 2*height/3);
  rotate(-PI/2);
  turtle.render();
  noLoop();
}

int counter = 0;

void mousePressed() {
  if (counter < 10) {
    pushMatrix();
    lsys.generate();
    //println(lsys.getSentence());
    turtle.setToDo(lsys.getSentence());
    turtle.changeLen(0.5); //para factor beta
    popMatrix();
    redraw();
    counter++;
  }
}
