package valentin;

import java.util.Scanner;

public class Ejercicio6 {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);

        int num = sc.nextInt();

        sc.close();

        int num_resp = 0;
        int num_div = 0;

        for(int i = 1; i < num; i++){

            int num_for_div = 0;

            for(int j = 1; j < i+1; j++){

                if(i % j == 0){

                    num_for_div+=1;

                }

            }

            if(num_for_div >= num_div){

                num_div = num_for_div;
                num_resp = i;

            }

        }

        System.out.println("El numero con mas divisores es "+ num_resp + " (tiene "+num_div + " divisores)");
    }
}
