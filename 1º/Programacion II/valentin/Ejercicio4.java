package valentin;

import java.util.Scanner;

class Ejercicio4{

    public static void main(String[] args) {

        Scanner scanner = new Scanner(System.in);
    
        int number = scanner.nextInt();

        scanner.close();

        String resp = "";


        for(int i = 1; i < number ; i++){

            Boolean resp_for = true;

            if(i > 1 && i % 1 == 0 && i % i == 0){

                for(int j = 2; j < i; j++){
            
                    if(i % j == 0){
                        
                        resp_for = false;
                        break;
                
                    }
                
                }

                if(resp_for){

                    resp += i +" ";

                }
            }

        }

        System.out.println(resp);
                
    }

}