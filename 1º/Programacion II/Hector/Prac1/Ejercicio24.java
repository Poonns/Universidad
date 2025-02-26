package Hector.Prac1;

import java.io.File;
import java.io.FileNotFoundException;
import java.util.Scanner;

public class Ejercicio24 {

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
    
}
