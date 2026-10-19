/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package themes;

import com.formdev.flatlaf.*;

/**
 *
 * @author Carla Terol
 */
public class PatitasLaf extends FlatLightLaf  {
    public static boolean setup() {
        
        return setup(new PatitasLaf());
    }

    @Override
    public String getName() {
        return "PatitasLaf";
    }
}

