package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio7 {

    public static boolean esPrimo(int n){

        if(n < 2){
            return false;
        }

        for(int i = 2; i < n; i++){
            if(n % i == 0){
                return false;

            }

        }
        
        return true;
    }
    
    public static void main(String[] args) {
        
        Scanner entrada = new Scanner(System.in);
        System.out.println("Introduc un numero: ");
        int n = entrada.nextInt();
        entrada.close();


        if(esPrimo(n)){
            System.out.println(n +" es primo");
        } else {
            System.out.println(n + " no es primo");
        }

    }

}
