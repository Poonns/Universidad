package valentin;

import java.util.Scanner;

public class Ejercicio16 {

    public static int olasDeFrio(double[] temperaturas, int numero){

        int cont = 0;

        int resp = 0;

        for(int i = 0; i < temperaturas.length;i++){

            if(temperaturas[i] < 0){

                cont+=1;

            }else if(temperaturas[i] > 0){
               cont = 0;
            }

            if(cont == numero){

                resp+=1;
            }

        }

        return resp;
    }


    public static void main(String[] args) {

        Scanner sc = new Scanner(System.in);

        System.out.println("Numero de dias: ");

        int numero =  sc.nextInt();

        sc.close();


        double[] temperaturas = {3.2, 1.5, -2.1, -3.5, -4.0, -1.8, 0.5, 2.0, -1.2, -2.5, -3.0, -4.1, -5.3, 0.2, 1.8, -6.0, -7.2, -3.5, -1.0, 2.3};


        System.out.println(olasDeFrio(temperaturas, numero));

    }
    
}
