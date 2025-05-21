package Hector.Examenes finales.Junio2022;

import java.io.FileNotFoundException;
import java.util.Scanner;
import java.io.File;

public class problema2 {
        public class Donacion {
        private String dni;
        private double importe;

        public Donacion(String dni, double importe) {
            this.dni = dni;
            this.importe = importe;
        }

        public String getDni() {
            return dni;
        }

        public double getImporte() {
            return importe;
        }
    }

    public class RegistroDonaciones {
        private static class Nodo {
            Donacion donacion;
            Nodo sig;

            Nodo(Donacion laDonacion, Nodo elSiguiente) {
                this.donacion = laDonacion;
                this.sig = elSiguiente;
            }
        }

        // Atributos:
        private Nodo primero;
        private Nodo ultimo;
        private double importeTotal;

        public RegistroDonaciones() {}

        void añadir(String dni, double importe){
            Nodo nuevo = new Nodo(new Donacion(dni, importe), null);
            if(primero == null){
                primero = nuevo;
            } else{
                ultimo.sig = nuevo;
            }
            ultimo = nuevo;
        }

        RegistroDonaciones(String fichero) throws FileNotFoundException{
            Scanner entrada = new Scanner(new File(fichero));

            while(entrada.hasNextLine()){
                String dni = entrada.next();
                double importe = entrada.nextDouble();
                añadir(dni, importe);
            }
        }

        RegistroDonaciones extrarDonacionesPremium(int importePremium){
            RegistroDonaciones nuevo = new RegistroDonaciones();
            Nodo anterior = null;
            for(Nodo actual = primero; actual != null;){
                if(actual.donacion.getImporte() >= (double) importePremium){
                    nuevo.añadir(actual.donacion.getDni(), actual.donacion.getImporte());
                    if(anterior == null){
                        primero = actual.sig;
                        actual = actual.sig;
                    } else{
                        anterior.sig = actual.sig;
                        actual = actual.sig;
                    }
                    importeTotal -= actual.donacion.getImporte();
                } else{
                    anterior = actual;
                    actual = actual.sig;
                } 
            }

            return nuevo;
        }


    }
}
