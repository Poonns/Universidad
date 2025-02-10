package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio10 {
    
    public static int contarDivisores(int n){
        
        int divisores = 0;
        
        for(int i = 1; i <= n; i++){
            if(n % i == 0){

                divisores ++;

            }
        }
        return divisores;
    }

    public static void main(String[] args) {

        Scanner entrada = new Scanner(System.in);
        int n = entrada.nextInt();
        entrada.close();
        int maxdivisores = 0;
        int mayor = 1;

        for(int j = 1; j <= n; j++){

            if(contarDivisores(j) > maxdivisores){
                maxdivisores = contarDivisores(j);
                mayor = j;

            }

        }
        
        
        System.out.println("El numero con mas divisores es " + mayor + " (" + maxdivisores + " divisores)");

    }

    
}
