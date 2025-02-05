package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio2 {

    public static void main(String[] args) {
        Scanner entrada = new Scanner(System.in);
        int n = entrada.nextInt();
        int productorio = 1;
        entrada.close();

        if(n == 1 || n == 0){

        

        } else if(n % 2 == 0){

            for(int i = 2; i < n+1; i += 2){

                productorio = productorio *= i;

            }
            
        } else if(n % 2 != 0) {

            for(int i = 1; i< n+1; i += 2){

                productorio = productorio * i;

            }
            
        }

        System.out.println(productorio);

    }



}
    

