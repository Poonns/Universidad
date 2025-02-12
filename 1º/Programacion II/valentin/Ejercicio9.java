package valentin;

import java.util.Scanner;

public class Ejercicio9 {

    public static int elPrimorial(int num){

        int num_resp = 1;

        for(int i = 2; i < num+1; i++){

            Boolean primo= true;

            for(int j = 2; j < i; j++){

                if(i % j == 0 ){
                    primo = false;

                    break;
                }
    
            }

            if(primo){

                num_resp= num_resp * i;

            }

        }

        return num_resp;
        

    }


    public static void main(String[] args) {
        
    
        Scanner sc = new Scanner(System.in);
    
        int num = sc.nextInt();
    
        sc.close();
    
        System.out.println(elPrimorial(num));
    }
    
}
