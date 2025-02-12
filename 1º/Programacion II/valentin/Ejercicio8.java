package valentin;

import java.util.Scanner;

public class Ejercicio8 {

    static String esPrimo(int num){


        String resp = "";

        if(num <= 1){

            return "No es primo";

        }else{

            for(int i = 2; i < num+1; i++){

                Boolean primo= true;

                for(int j = 2; j < i; j++){

                    if(i % j == 0 ){
                        primo = false;

                        break;
                    }
        
                }

                if(primo){

                    resp+= i + " ";

                }
    
            }

            return resp;

        }

    }
    


    public static void main(String[] args) {
        
    
        Scanner sc = new Scanner(System.in);
    
        int num = sc.nextInt();
    
        sc.close();
    
        System.out.println(esPrimo(num));
    }
}
