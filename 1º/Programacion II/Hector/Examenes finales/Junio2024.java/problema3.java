package Hector.Examenes finales.Junio2024.java;

public class problema3 {
    int contarDeportistasAN(Deportista[] deportistas){
        int contador = 0;
        int j = 0;
        Nodo actual = primero;

        while(actual != null && j < deportistas.length){
            if(actual.estudiante.getdni().compareTo(deportistas[j].getDni()) == 0){
                contador++;
                j++;
                actual = actual.sig;
            } else if(actual.estudiante.getdni().compareTo(deportistas[j].getDni()) < 0){
                actual = actual.sig;
            } else{
                j++;
            }
        }

        return contador;
    }
}
