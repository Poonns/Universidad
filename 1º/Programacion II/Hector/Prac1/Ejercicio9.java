package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio9 {
    
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
            System.out.println("Introduce un numero entero: ");
            int n = entrada.nextInt();
            entrada.close();
            int primoral = 1;

            for(int i = 2; i <= n; i++){

                if(esPrimo(i)){

                    primoral *= i;
                }

            }

            System.out.println("El primoral de " + n + " es " + primoral);

        }

}
