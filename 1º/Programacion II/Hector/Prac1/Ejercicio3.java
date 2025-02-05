package Hector;

import java.util.Scanner;


public class Ejercicio3 {
    
    public static void main(String[] args){

        Scanner entrada = new Scanner(System.in);
        int n = entrada.nextInt();
        entrada.close();
        
        if(n < 2){

            System.out.println(n + "o es primo");
            return;

        } 
        
        boolean esPrimo = true;
        
        for(int i = 2; i < n+1 ; i++){

            if(n % i == 0){

                esPrimo = false;
                break;

            }

        }

        if(esPrimo){

            System.err.println(n + "es primo");

        } else {

            System.err.println(n + "no es primo");

        }

    }

}
