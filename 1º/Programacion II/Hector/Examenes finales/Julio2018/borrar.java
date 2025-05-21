package Hector.Examenes

import java.io.FileNotFoundException;
import java.util.Scanner;
import java.io.File;


public class borrar {
    class Donacion{
        private String dni;
        private double importe;

        Donacion(String dni, double importe){
            this.dni = dni;
            this.importe = importe;
        }

        String getDni(){
            return dni;
        }

        double getImporte(){
            return importe;
        }

        String toString(){
            return dni + ": " + importe +"€";
        }
    }

    public class RegistroDonaciones {
    private class Nodo {
        Donacion donacion;
        Nodo sig;

        Nodo(Donacion laDonacion, Nodo elSiguiente) {
            this.donacion = laDonacion;
            this.sig = elSiguiente;
        }
    }

    private Nodo primero;
    private Nodo ultimo;
    private double importeTotal;

    public RegistroDonaciones(){}

    void añadir(String dni, double importe){
        Nodo nuevo = new Nodo(new Donacion(dni, importe), null);
        if(primero == null){
            primero = nuevo;
            ultimo = nuevo;
        } else{
            ultimo.sig = nuevo;
            ultimo = nuevo;
        }
    }

    RegistroDonaciones(String fichero) throws FileNotFoundException{

        Scanner entrada = new Scanner(new File(fichero));

        while(entrada.hasNextLine()){
            String dni = entrada.next();
            double importe = entrada.nextDouble();
            añadir(dni, importe);

        }
    }

    RegistroDonaciones extraerDonacionesPremium(int importePremium){
        int contador = 0;
        for(Nodo actual = primero; actual != null; actual = actual.sig){
            if(actual.donacion.importe > importePremium){
                contador ++;
            }
        }

        Donacion[] nuevo = new Donacion[contador];
        int i = 0;
        for(Nodo actual = primero; actual != null; actual = actual.sig){
            if(actual.donacion.importe > importePremium){
                nuevo[i] = 
            }
        }
    }
    
}
