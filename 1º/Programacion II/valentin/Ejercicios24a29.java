package valentin; // Indica aquí el nombre del paquete que estás utilizando

import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.ObjectInputStream;
import java.util.Arrays;
import java.util.Scanner;

public class Ejercicios24a29 {

    public static void main(String[] args) throws FileNotFoundException {
        String[] elementos10 = {"tierra", "fuego", "lava", "aire", "piedra", "agua", "arena", "desierto", "cristal", "reloj_de_arena"};
        String nombre10 = "1º/Programacion II/archivos/alquimia10.txt";
        int[][] matriz10 = leerMatrizResultado("1º/Programacion II/archivos/matriz10.bin");
        String[] elementos380 = leerElementos("1º/Programacion II/archivos/elementos380.txt");
        String nombre380 = "1º/Programacion II/archivos/alquimia380.txt";
        int[][] matriz380 = leerMatrizResultado("1º/Programacion II/archivos/matriz380.bin");

        System.out.println("*** PROBANDO CON LOS ARCHIVOS DE 10 ELEMENTOS ***");
        prueba(elementos10, nombre10, matriz10, new String[]{null, "desierto", "reloj_de_arena", null},
                new String[]{"arena"},
                2, "arena", 1);

        System.out.println("\n******************************\n");

        System.out.println("*** PROBANDO CON LOS ARCHIVOS DE 380 ELEMENTOS ***");
        prueba(elementos380, nombre380, matriz380, new String[]{"viento", "desierto", "reloj_de_arena", null},
                new String[]{"agua", "aire", "árbol", "arboleda", "arena", "ciudad", "continente", "cuchillo", "escarabajo", "libro", "país", "rascacielos", "rueda", "tierra", "transistor", "tumba"},
                154,"fuego", 0);
    }

    public static void prueba(String[] elementos, String nFichComb, int[][] matRes,
                              String[] elemEsperados, String[] combConsigoEsperado,
                              int nTerminalesEsperados, String másCombinableEsperado,
                              int nMúltiplesCombEsperado) throws FileNotFoundException {
        final String ANSI_RED = "\u001B[31m";
        final String ANSI_GREEN = "\u001B[32m";
        final String ANSI_RESET = "\u001B[0m";

        System.out.println("CREANDO LA MATRIZ DE COMBINACIONES");
        int[][] m = crearMatrizCombinaciones(elementos, nFichComb);
        if (Arrays.deepEquals(m, matRes))
            System.out.println(ANSI_GREEN + "La matriz construida es correcta");
        else
            System.out.println(ANSI_RED + "La matriz obtenida no coincide con la esperada");
        System.out.print(ANSI_RESET);
        /* ----------------------------- */
        System.out.println("PROBANDO El MÉTODO ELEMENTO_GENERADO");
        String[][] combo = {{"aire", "aire"}, {"arena", "arena"}, {"arena", "cristal"}, {"aire", "xyz012"}};
        for (int i = 0; i < elemEsperados.length; i++) {
            String elem1 = combo[i][0];
            String elem2 = combo[i][1];
            String elementoObtenido = elementoGenerado(m, elementos, elem1, elem2);
            System.out.format("Al combinar %s y %s se produce %s", elem1, elem2, elementoObtenido);
            if (elementoObtenido == elemEsperados[i] || (elementoObtenido != null && elementoObtenido.equals(elemEsperados[i])))
                System.out.println(" --> " + ANSI_GREEN + "OK" + ANSI_RESET);
            else
                System.out.println(" --> " + ANSI_RED + "Error. Se esperaba: " + elemEsperados[i] + ANSI_RESET);
        }
        /* ----------------------------- */
        System.out.println("PROBANDO EL MÉTODO COMBINABLES_CONSIGO");
        String[] resCombConsigo = combinablesConsigo(m, elementos);
        System.out.format("Se obtiene(n) %d elemento(s) combinable(s) consigo: %s%n ", resCombConsigo.length, Arrays.toString(resCombConsigo));
        if (Arrays.equals(resCombConsigo, combConsigoEsperado))
            System.out.println("\t--> " + ANSI_GREEN + "OK" + ANSI_RESET);
        else
            System.out.println("\t--> " + ANSI_RED + "Error. Se esperaba: " + Arrays.toString(combConsigoEsperado) + ANSI_RESET);
        /* ----------------------------- */
        System.out.println("PROBANDO EL MÉTODO CONTAR_TERMINALES");
        int nTerminales = contarTerminales(m);
        System.out.format("Se obtienen %d elementos terminales", nTerminales);
        if (nTerminales == nTerminalesEsperados)
            System.out.println(" --> " + ANSI_GREEN + "OK" + ANSI_RESET);
        else
            System.out.println(" --> " + ANSI_RED + "Error. Se esperaban: " + nTerminalesEsperados + ANSI_RESET);
        /*  ----------------------------- */
        System.out.println("PROBANDO EL MÉTODO ELEMENTO_MÁS_COMBINABLE");
        String másCombinable = elementoMásCombinable(m, elementos);
        System.out.format("Se obtiene %s como elemento más combinable", másCombinable);
        if (másCombinable.equals(másCombinableEsperado))
            System.out.println(" --> " + ANSI_GREEN + "OK" + ANSI_RESET);
        else
            System.out.println(" --> " + ANSI_RED + "Error. Se esperaba: " + másCombinableEsperado + ANSI_RESET);
        // /* ----------------------------- */
        System.out.println("PROBANDO EL MÉTODO MÚLTIPLES_COMBINACIONES");
        int nMúltiplesComb = múltiplesCombinaciones(m);
        System.out.format("%d elemento(s) se puede(n) obtener con múltiples combinaciones", nMúltiplesComb);
        if (nMúltiplesComb == nMúltiplesCombEsperado)
            System.out.println(" --> " + ANSI_GREEN + "OK" + ANSI_RESET);
        else
            System.out.println(" --> " + ANSI_RED + "Error. Se esperaban: " + nMúltiplesCombEsperado + ANSI_RESET);

    }

    public static String[] leerElementos(String nFich) throws FileNotFoundException {
        Scanner entrada = new Scanner(new File(nFich));
        int nElem = entrada.nextInt();
        String[] vector = new String[nElem]; // Al comienzo del fichero se indica la cantidad de elementos
        for (int i = 0; i < nElem; i++)
            vector[i] = entrada.next();
        entrada.close();
        return vector;
    }

    public static int[][] leerMatrizResultado(String nFich) {
        try {
            ObjectInputStream in = new ObjectInputStream(new FileInputStream(nFich));
            int[][] m = (int[][]) in.readObject();
            in.close();
            return m;
        } catch (Exception e) {
            System.out.println("Algo ha ido mal al leer el fichero --> " + e.getMessage());
            System.exit(-1);
        }
        return null;
    }

    public static int[][] crearMatrizCombinaciones(String[] elementos, String nFich) throws FileNotFoundException {
        int[][] datos = new int[elementos.length][elementos.length];

        for(int i = 0; i < elementos.length;i++){
            for(int j = 0; j < elementos.length;j++){

                datos[i][j] = -1;

            }

        }

        Scanner sc = new Scanner(new File(nFich));

        while (sc.hasNextLine()) {
            String combinaciones_string = sc.nextLine();
            String[] data = combinaciones_string.split(" +"); 

            int a = 0;
            int b = 0;
            int c = 0;

            for(int i = 0; i < elementos.length ;i++){

                if(elementos[i].equals(data[0]) && elementos[i].equals(data[1])){
                    a = i;
                    b = i;
                }else if(elementos[i].equals(data[0])){
                    a = i;
                }else if(elementos[i].equals(data[1])){
                    b = i;
                }else if(elementos[i].equals(data[2])){
                    c = i;
                }
            }

            datos[a][b] = c;
            datos[b][a] = c;
            
        }
        
        sc.close();

        return datos; // Devuelve el resultado adecuado
    }

    public static String elementoGenerado(int[][] m, String[] elementos, String e1, String e2) {

        int index_A = 0;
        int index_B = 0;

        for(int i = 0; i< elementos.length;i++){
            if(elementos[i].equals(e1) && elementos[i].equals(e2) ){
                index_A = i;
                index_B = i;
                break;
            }else if(elementos[i].equals(e1)){
                index_A = i;
            }else if(elementos[i].equals(e2)){
                index_B = i;
            }
        }
        
        if(m[index_A][index_B] == -1){
            return null; 
        }
        
        return elementos[m[index_A][index_B]];
    }

    public static String[] combinablesConsigo(int[][] m, String[] elementos) {


        String[] vector =  new String[0];

        for(int i = 0; i < elementos.length;i++){

            String[] nuevo_vector = new String[vector.length+1];

            if(m[i][i] != -1){
                for(int j = 0; j< vector.length;j++){
                    nuevo_vector[j] = vector[j];
                }

                nuevo_vector[nuevo_vector.length-1] = elementos[i];
                vector = nuevo_vector;
            }



        } 

        return vector; // Devuelve el resultado adecuado
    }

    public static int contarTerminales(int[][] m) {
        int resp = 0;
        for(int i = 0;i < m.length;i++){
            boolean terminal = true;

            for(int num: m[i]){

                if(num != -1){
                    terminal = false;
                }

            }

            if(terminal){
                resp++;
            }
        }
        return resp; // Devuelve el resultado adecuado
    }

    public static String elementoMásCombinable(int[][] m, String[] elementos) {
        String resp = "";
        int cantidad_combinable = 0;
        for(int i = 0; i < elementos.length;i++){
            int cont = 0;
            for(int num: m[i]){

                if(num != -1){

                    cont++;

                }

            }

            if(cont > cantidad_combinable){

                cantidad_combinable = cont;
                resp = elementos[i];

            }

        }
        return resp; // Devuelve el resultado adecuado
    }

    public static int múltiplesCombinaciones(int[][] m) {

        int combinaciones_multiples = 0;
        
        for(int i = 0; i < m.length;i++){

            String datos = "";

            for(int j = 0; j < m.length;j++){

                for(int k = 0; k < m.length;k++){

                    if(m[j][k] == i || m[k][j] == i){
                        datos = datos + j +"-"+k+",";
                    }
                }

            }

            String[] datos_split = datos.split(",+");



            if(datos_split.length > 2){
                combinaciones_multiples++;
            }


        }
        return combinaciones_multiples; // Devuelve el resultado adecuado
    }
}
