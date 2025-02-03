package valentin;

public class Ejercicio2 {

    public static void main(String[] args){

        int num = 10;

        int res = 1;


        if(num == 1 || num == 0){

            res = 1;

        }else if(num % 2 == 0){

            for(int i = 2; i < num+1; i+=2){

                res = res * i;

            }
            
        }else{
            
            for(int i = 1; i < num+1; i+=2){

                res = res * i;

            }

        }

        System.out.println(res);

    }
    
}
