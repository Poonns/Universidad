package valentin;

import java.util.Scanner;

class Ejercicio3{


    public static void main(String[] args) {
        
        Scanner scanner = new Scanner(System.in);

        int number = scanner.nextInt();

        scanner.close();

        if(number < 2 || number % 1 == 0 && number % number == 0){
        
            Boolean resp = true;
            
            for(int i = 2; i < number; i++){
            
                if(number % i == 0){
            
                    resp = false;
            
                    break;
            
                }
            
            }

            if(resp){
            
                System.out.println("Es primo");
                        
            }else{
            
                System.out.println("No es primo");
            
            }

        }else{

            System.out.println("No es primo");
        }

    }

}