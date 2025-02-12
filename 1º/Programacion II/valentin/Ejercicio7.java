package valentin;

import java.util.Scanner;

public class Ejercicio7 {


    static Boolean esPrimo(int num){

        Boolean data = true;


        for(int i = 2; i < num; i++){

            if(num % i == 0){
                data = false;
            }

        }

        return data;

    }

    public static void main(String[] args) {

        Scanner sc = new Scanner(System.in);

        int num = sc.nextInt();

        sc.close();

        Boolean resp  = esPrimo(num);


        if(!resp){

            System.out.println("No es primo");
            
        }else{

            System.out.println("Es primo");

        }
    }
    
}
