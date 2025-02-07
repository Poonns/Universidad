package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio5 {

    public static void main(String[] args) {
        
        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduce un numero: ");
        int n = entrada.nextInt();
        entrada.close();
        int primorial = 1;

        for(int i = 2; i <= n ; i++){

            int divisores = 0;

            for(int j = 1; j <= i; j++){

                if(i % j == 0){

                    divisores ++ ; 

                }

            }

            if(divisores == 2){

                primorial *= i ;

            }

        }

        System.out.println("El primorial es " + primorial);

    }
    
}
