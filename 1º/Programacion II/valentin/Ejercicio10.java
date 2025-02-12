package valentin;

import java.util.Scanner;

public class Ejercicio10 {

    public static String contarDivisores(int num){

        int num_div = 0;

        for(int i = 1; i < num+1; i++){


                if(num % i == 0){

                    num_div+=1;
                }

        }



        return "El numero "+ num + " tiene "+num_div + " divisores";

    }

    
    public static void main(String[] args) {

        Scanner sc = new Scanner(System.in);

        int num = sc.nextInt();

        sc.close();

        System.out.println(contarDivisores(num));
    }
}
