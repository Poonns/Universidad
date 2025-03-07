package Hector.Prac1;

import java.util.Scanner;

public class Ejercicio13 {

    public static String obtenerPalabra(String cadena, int n) {
        int contador = 0;

        for (int i = 0; i < cadena.length(); i++) {

            char caracteres_cadena = cadena.charAt(i);

            if (caracteres_cadena != ' ' && (i == 0 || cadena.charAt(i - 1) == ' ')) {
                contador++;

                if (contador == n) {
                    int inicio = i;
                    int fin = i;

                    for (int j = i; j < cadena.length(); j++) {
                        if (cadena.charAt(j) == ' ') {
                            break;
                        }
                        fin = j;
                    }

                    return cadena.substring(inicio, fin + 1);
                }
            }
        }

        return null;
    }

    public static void main(String[] args) {
        Scanner scanner = new Scanner(System.in);
        System.out.print("Ingresa una cadena: ");
        String texto = scanner.nextLine();
        System.out.print("Ingresa el número de la palabra a extraer: ");
        int n = scanner.nextInt();
        scanner.close();

        String palabra = obtenerPalabra(texto, n);
        System.out.println("Palabra encontrada: " + palabra);
    }
}
