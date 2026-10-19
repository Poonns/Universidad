import java.io.*;

%%
%class ConvertirAMayusculas
%unicode
%standalone

%{ 
    FileWriter escritor;
%}

%init{
    try{
        escritor = new FileWriter("salida.may");
    } catch(Exception e){
        e.printStackTrace();
    }
%init}

%eof{
    try {
        if (escritor != null) escritor.close();
    } catch (Exception e) {
        e.printStackTrace();
    }
%eof}

%%
[a-zñáéíóú] { escritor.write(yytext().toUpperCase()); }
[^a-zñáéíóú] { escritor.write(yytext()); }
