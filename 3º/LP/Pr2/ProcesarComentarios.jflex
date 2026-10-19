import java.io.*;

%%
%class ProcesarComentarios
%unicode
%standalone

%{ 
    FileWriter cod;
    FileWriter com;
    int charCodigo = 0;
    int charComentarios = 0;
%}

%init{
    try {
        cod = new FileWriter("salida.cod");
        com = new FileWriter("salida.com");
    } catch(Exception e) {
        e.printStackTrace();
    }
%init}

%eof{
    try {
        if (cod != null) cod.close();
        if (com != null) com.close();
    } catch (Exception e) {
        e.printStackTrace();
    }

    double porcentaje = (charComentarios * 100.0) / (charCodigo + charComentarios);
    System.out.println("Porcentaje de comentarios: " + porcentaje + "%");
%eof}

%%

// 1. Comentarios de línea
"//"[^\n\r]* { 
    com.write(yytext()); 
    charComentarios += yytext().length(); 
}

"/*" ~"*/" { 
    com.write(yytext()); 
    charComentarios += yytext().length(); 
}

\n | \r | \r\n { 
    cod.write(yytext());
    com.write(yytext());
}

. {
    cod.write(yytext());
    charCodigo += yytext().length();
}