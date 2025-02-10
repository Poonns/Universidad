package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio6 {

    public static void main(String[] args) {

        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduce un número entero: ");
        int n = entrada.nextInt();
        entrada.close();
        
        int mayor = 1;
        int maxdivisores = 1;

        for(int i = 1; i <= n; i++){

            int divisores = 0;

            for(int j = 1; j <= i; j++){

                if(i % j == 0){

                    divisores ++ ; 

                }      

            }

            if(divisores > maxdivisores){

                maxdivisores = divisores;
                mayor = i;

            }

        }

        System.out.println("El numero con mas divisores es " + mayor + " (" + maxdivisores + " divisores)");
        
    }
    
}
