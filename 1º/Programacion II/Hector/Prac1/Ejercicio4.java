package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio4 {

    public static void main(String[] args){

        Scanner entrada = new Scanner(System.in);
        int n = entrada.nextInt();
        entrada.close();

        for(int i = 2; i < n ; i++){

            int divisores = 0;

            for(int j = 1; j <= i; j++){

                if(i % j == 0){

                    divisores ++;

                }

            }

            if(divisores == 2){

                System.out.println(i);

            }

        }

    }
    
}
