%%

%class AnalizadorJFlex
%unicode
%standalone

LETRA   = [a-zA-Z]
DIGITO  = [0-9]
ID      = {LETRA}{1,8}
NUMERO  = {DIGITO}+
ESPACIO = [ \t\r\n]

%state COMENTARIO

%{
    private int profundidad = 0;
%}

%%

<YYINITIAL> {
    "leer" { System.out.println("PALABRA CLAVE LEER"); }
    "mostrar" { System.out.println("PALABRA CLAVE MOSTRAR"); }
    {ID} { System.out.println("IDENTIFICADOR: " + yytext()); }
    {NUMERO} { System.out.println("NÚMERO: " + yytext()); }
    ";" { System.out.println("PUNTO Y COMA"); }
    "<-" { System.out.println("SÍMBOLO DE ASIGNACIÓN"); }
    "+" { System.out.println("OPERADOR DE SUMA"); }
    "-" { System.out.println("OPERADOR DE RESTA"); }
    "*" { System.out.println("OPERADOR DE MULTIPLICACIÓN"); }
    "/"  { System.out.println("OPERADOR DE DIVISIÓN"); }
    "//".* { /* comentario de línea: se ignora */ }
    "/*" { profundidad = 1; yybegin(COMENTARIO); }
    {ESPACIO}   { /* se ignora */ }
    <<EOF>>         { System.out.println("--- FIN DE CADENA ---"); return 0; }
}

<COMENTARIO> {
    "/*" { profundidad++; }
    "*/" { profundidad--; if (profundidad == 0) yybegin(YYINITIAL); }
    [^]  { /* cualquier otro carácter dentro del comentario: se ignora */ }
}
