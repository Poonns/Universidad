/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package com.mycompany.practica3;

/**
 *
 * @author hectorpons
 */
public class Analizador {

    public static void analizar(String cadena) {
        int i = 0;
        int n = cadena.length();

        while (i < n) {
            char c = cadena.charAt(i);

            if (Character.isWhitespace(c)) {
                i++;
                continue;
            }

            if (Character.isLetter(c)) {
                StringBuilder sb = new StringBuilder();
                while (i < n && Character.isLetter(cadena.charAt(i))) {
                    sb.append(cadena.charAt(i));
                    i++;
                }
                String palabra = sb.toString();
                
                // CORRECCIÓN: Limitar el identificador a un máximo de 8 caracteres según el enunciado
                if (palabra.length() > 8) {
                    palabra = palabra.substring(0, 8);
                }

                if (palabra.equals("leer")) {
                    System.out.println("PALABRA CLAVE LEER");
                } else if (palabra.equals("mostrar")) {
                    System.out.println("PALABRA CLAVE MOSTRAR");
                } else {
                    System.out.println("IDENTIFICADOR: " + palabra);
                }
                continue;
            }

            if (Character.isDigit(c)) {
                StringBuilder sb = new StringBuilder();
                while (i < n && Character.isDigit(cadena.charAt(i))) {
                    sb.append(cadena.charAt(i));
                    i++;
                }
                System.out.println("NÚMERO: " + sb.toString());
                continue;
            }

            switch (c) {
                case ';':
                    System.out.println("PUNTO Y COMA");
                    i++;
                    continue;
                case '+':
                    System.out.println("OPERADOR DE SUMA");
                    i++;
                    continue;
                case '-':
                    System.out.println("OPERADOR DE RESTA");
                    i++;
                    continue;
                case '*':
                    System.out.println("OPERADOR DE MULTIPLICACIÓN");
                    i++;
                    continue;
                case '/':
                    System.out.println("OPERADOR DE DIVISIÓN");
                    i++;
                    continue;
                case '<':
                    if (i + 1 < n && cadena.charAt(i + 1) == '-') {
                        System.out.println("SÍMBOLO DE ASIGNACIÓN");
                        i += 2;
                        continue;
                    }
                    break;
            }

            // carácter no reconocido: se salta
            i++;
        }

        System.out.println("--- FIN DE CADENA ---");
    }

    public static void main(String[] args) {
        // Prueba con la cadena de ejemplo de la práctica
        analizar("leer x;\ny <- 2*x;\nmostrar y;");
    }
}