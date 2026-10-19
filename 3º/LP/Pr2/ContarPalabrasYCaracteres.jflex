import java.io.*;

%%
%class ContarPalabrasYCaracteres
%unicode
%standalone

%{
    int lineas = 0;
    int palabras = 0;
    int numeros = 0;
    int caracteres = 0;
    int caracteresSinEspacios = 0;
%}

%eof{
  System.out.println("Número de líneas: " + lineas);
  System.out.println("Número de palabras: " + palabras);
  System.out.println("Número de cantidades numéricas: " + numeros);
  System.out.println("Número de caracteres totales: " + caracteres);
  System.out.println("Número de caracteres sin espacios/saltos: " + caracteresSinEspacios);
%eof}

%%
\n | \r | \r\n { 
    lineas++; 
    caracteres += yytext().length(); 
}

[a-zA-ZñÑáéíóúÁÉÍÓÚ][a-zA-Z0-9ñÑáéíóúÁÉÍÓÚ]* { 
    palabras++; 
    caracteres += yytext().length(); 
    caracteresSinEspacios += yytext().length(); 
}

[0-9]+ { 
    numeros++;
    caracteres += yytext().length(); 
    caracteresSinEspacios += yytext().length(); 
}

[ \t]+ { 
    caracteres += yytext().length(); 
}

. { 
    caracteres += yytext().length(); 
    caracteresSinEspacios += yytext().length(); 
}
