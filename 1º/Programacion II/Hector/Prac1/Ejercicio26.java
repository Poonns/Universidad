package Hector.Prac1;

import java.io.File;
import java.io.FileNotFoundException;
import java.util.Scanner;

public class Ejercicio26 {

    public static int[][] crearMatrizCombinaciones(String[] elementos, String fichero) throws FileNotFoundException {

        int N = elementos.length;
        int[][] matriz = new int[N][N];

        for(int i = 0; i < N; i++){

            for(int j = 0; j < N; j++){ 

                matriz[i][j] = -1;

            }

        }

        Scanner entrada = new Scanner(new File(fichero));

        while (entrada.hasNextLine()){

            String[] partes = entrada.nextLine().split(" ");

            if(partes.length == 3){

                String elemento1 = partes[0];
                String elemento2 = partes[1];
                String resultado = partes[2];

                int i = -1 , j = -1, k = -1;
                
                for(int x = 0; x < N; x++){

                    if(elementos[x].equals(elemento1)){
                        i = x;
                    }

                    if(elementos[x].equals(elemento2)){
                        j = x;
                    }

                    if(elementos[x].equals(resultado)){
                        k = x;
                    }
                
                }

                if(i != -1 && j != -1 && k != -1){

                    matriz[i][j] = k;
                    matriz[j][i] = k;

                }

            }

        }

        entrada.close();
        return matriz;

    }

    public static String elementoGenerado(int[][] matriz, String[] elementos, String cadena1, String cadena2){

        int i = -1, j = -1;
        
        for(int x = 0; x < elementos.length; x++){

            if(elementos[x].equals(cadena1)){

                i = x;

            }

            if(elementos[x].equals(cadena2)){

                j = x;

            }

        }

        if(i == -1 || j == -1){

            return null;

        }

        int resultado_matriz = matriz[i][j];

        if(resultado_matriz != -1){

            return elementos[resultado_matriz];

        } else {

            return null;

        }

    }

    public static String[] combinablesConsigo(int[][] matriz, String[] elementos){

        int contador = 0;

        for(int i = 0; i < elementos.length; i++){
            
            if(matriz[i][i] != -1){

                contador ++ ;

            }

        }

        String[] combinables = new String[contador];
        int indice = 0;

        for(int i = 0; i < elementos.length; i++){

            if(matriz[i][i] != -1){

                combinables[indice] = elementos[i];
                indice ++ ;

            }

        }

        return combinables;

    }

}

